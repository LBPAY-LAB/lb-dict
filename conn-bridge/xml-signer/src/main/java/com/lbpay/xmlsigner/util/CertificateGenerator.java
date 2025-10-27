package com.lbpay.xmlsigner.util;

import org.bouncycastle.asn1.x500.X500Name;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.cert.X509CertificateHolder;
import org.bouncycastle.cert.X509v3CertificateBuilder;
import org.bouncycastle.cert.jcajce.JcaX509CertificateConverter;
import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.bouncycastle.operator.ContentSigner;
import org.bouncycastle.operator.jcajce.JcaContentSignerBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.FileOutputStream;
import java.math.BigInteger;
import java.security.*;
import java.security.cert.X509Certificate;
import java.util.Date;

/**
 * Utility class for generating self-signed certificates for development/testing
 * NOT FOR PRODUCTION USE - Only for dev mode with self-signed certificates
 */
public class CertificateGenerator {

    private static final Logger logger = LoggerFactory.getLogger(CertificateGenerator.class);

    static {
        Security.addProvider(new BouncyCastleProvider());
    }

    /**
     * Generate a self-signed certificate and save to PKCS12 file
     *
     * @param outputPath Path to save the .p12 file
     * @param password Password for the keystore
     * @param alias Key alias
     * @param commonName Certificate common name (CN)
     * @param validityDays Certificate validity in days
     */
    public static void generateSelfSignedCertificate(
            String outputPath,
            String password,
            String alias,
            String commonName,
            int validityDays) throws Exception {

        logger.info("Generating self-signed certificate: CN={}, validity={}days", commonName, validityDays);

        // Generate key pair
        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance("RSA", "BC");
        keyPairGenerator.initialize(2048, new SecureRandom());
        KeyPair keyPair = keyPairGenerator.generateKeyPair();

        // Certificate validity period
        Date startDate = new Date(System.currentTimeMillis() - 24 * 60 * 60 * 1000); // 1 day back
        Date endDate = new Date(System.currentTimeMillis() + validityDays * 24L * 60 * 60 * 1000);

        // Certificate serial number
        BigInteger serialNumber = new BigInteger(64, new SecureRandom());

        // Build certificate
        X500Name issuerName = new X500Name("CN=" + commonName + ", O=LBPay, C=BR");
        X500Name subjectName = issuerName;

        SubjectPublicKeyInfo subjectPublicKeyInfo = SubjectPublicKeyInfo.getInstance(
                keyPair.getPublic().getEncoded());

        X509v3CertificateBuilder certBuilder = new X509v3CertificateBuilder(
                issuerName,
                serialNumber,
                startDate,
                endDate,
                subjectName,
                subjectPublicKeyInfo
        );

        // Sign certificate
        ContentSigner contentSigner = new JcaContentSignerBuilder("SHA256WithRSAEncryption")
                .setProvider("BC")
                .build(keyPair.getPrivate());

        X509CertificateHolder certHolder = certBuilder.build(contentSigner);
        X509Certificate certificate = new JcaX509CertificateConverter()
                .setProvider("BC")
                .getCertificate(certHolder);

        // Create PKCS12 keystore
        KeyStore keyStore = KeyStore.getInstance("PKCS12");
        keyStore.load(null, null);

        // Add private key and certificate
        keyStore.setKeyEntry(
                alias,
                keyPair.getPrivate(),
                password.toCharArray(),
                new X509Certificate[]{certificate}
        );

        // Save to file
        try (FileOutputStream fos = new FileOutputStream(outputPath)) {
            keyStore.store(fos, password.toCharArray());
        }

        logger.info("Self-signed certificate saved to: {}", outputPath);
        logger.info("Certificate info - Subject: {}, Valid from: {} to: {}",
                certificate.getSubjectX500Principal().getName(),
                certificate.getNotBefore(),
                certificate.getNotAfter());
    }

    /**
     * Main method for standalone certificate generation
     */
    public static void main(String[] args) {
        try {
            String outputPath = args.length > 0 ? args[0] : "test-certificate.p12";
            String password = args.length > 1 ? args[1] : "changeit";
            String alias = args.length > 2 ? args[2] : "test";
            String cn = args.length > 3 ? args[3] : "LBPay Test Certificate";
            int validity = args.length > 4 ? Integer.parseInt(args[4]) : 365;

            generateSelfSignedCertificate(outputPath, password, alias, cn, validity);
            System.out.println("Certificate generated successfully!");
            System.out.println("File: " + outputPath);
            System.out.println("Password: " + password);
            System.out.println("Alias: " + alias);

        } catch (Exception e) {
            System.err.println("Failed to generate certificate: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }
}
