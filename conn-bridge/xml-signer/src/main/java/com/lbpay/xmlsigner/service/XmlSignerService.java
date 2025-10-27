package com.lbpay.xmlsigner.service;

import com.lbpay.xmlsigner.exception.XmlSignerException;
import com.lbpay.xmlsigner.model.SignRequest;
import com.lbpay.xmlsigner.model.SignResponse;
import org.apache.xml.security.signature.XMLSignature;
import org.apache.xml.security.transforms.Transforms;
import org.apache.xml.security.utils.Constants;
import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.xml.sax.InputSource;

import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerFactory;
import javax.xml.transform.dom.DOMSource;
import javax.xml.transform.stream.StreamResult;
import java.io.FileInputStream;
import java.io.StringReader;
import java.io.StringWriter;
import java.security.*;
import java.security.cert.X509Certificate;
import java.util.Enumeration;

/**
 * Service for XML digital signature operations using ICP-Brasil certificates
 */
@Service
public class XmlSignerService {

    private static final Logger logger = LoggerFactory.getLogger(XmlSignerService.class);

    static {
        // Register BouncyCastle provider for ICP-Brasil support
        Security.addProvider(new BouncyCastleProvider());
    }

    /**
     * Sign XML document with digital signature
     *
     * @param request Sign request containing XML and certificate info
     * @return Sign response with signed XML
     */
    public SignResponse signXml(SignRequest request) {
        try {
            logger.info("Starting XML signing process - DevMode: {}, CertPath: {}",
                    request.getDevMode(), request.getCertificatePath());

            // Load certificate and private key
            KeyStore keyStore = loadKeyStore(request);
            String alias = getKeyAlias(keyStore, request.getKeyAlias());
            PrivateKey privateKey = (PrivateKey) keyStore.getKey(alias,
                    request.getCertificatePassword() != null ?
                            request.getCertificatePassword().toCharArray() : null);
            X509Certificate certificate = (X509Certificate) keyStore.getCertificate(alias);

            if (privateKey == null || certificate == null) {
                throw new XmlSignerException("CERT_ERROR",
                        "Failed to load private key or certificate",
                        "Alias: " + alias);
            }

            // Parse XML document
            Document doc = parseXmlDocument(request.getXmlContent());

            // Sign the document
            signDocument(doc, privateKey, certificate, request);

            // Convert back to string
            String signedXml = documentToString(doc);

            // Get certificate info
            String certInfo = getCertificateInfo(certificate);

            logger.info("XML signed successfully with certificate: {}", certInfo);
            return SignResponse.success(signedXml, certInfo);

        } catch (XmlSignerException e) {
            logger.error("XML signing failed: {}", e.getMessage(), e);
            return SignResponse.error(e.getMessage());
        } catch (Exception e) {
            logger.error("Unexpected error during XML signing", e);
            return SignResponse.error("Unexpected error: " + e.getMessage());
        }
    }

    /**
     * Load KeyStore from file (PKCS12 or JKS)
     */
    private KeyStore loadKeyStore(SignRequest request) throws Exception {
        String certPath = request.getCertificatePath();
        String password = request.getCertificatePassword();

        KeyStore keyStore;
        try {
            // Try PKCS12 first (most common for ICP-Brasil A3)
            keyStore = KeyStore.getInstance("PKCS12");
            try (FileInputStream fis = new FileInputStream(certPath)) {
                keyStore.load(fis, password != null ? password.toCharArray() : null);
            }
            logger.info("Loaded PKCS12 keystore from: {}", certPath);
        } catch (Exception e) {
            // Fallback to JKS
            try {
                keyStore = KeyStore.getInstance("JKS");
                try (FileInputStream fis = new FileInputStream(certPath)) {
                    keyStore.load(fis, password != null ? password.toCharArray() : null);
                }
                logger.info("Loaded JKS keystore from: {}", certPath);
            } catch (Exception jksError) {
                throw new XmlSignerException("KEYSTORE_ERROR",
                        "Failed to load keystore as PKCS12 or JKS",
                        "Path: " + certPath);
            }
        }

        return keyStore;
    }

    /**
     * Get key alias from keystore
     */
    private String getKeyAlias(KeyStore keyStore, String preferredAlias) throws Exception {
        if (preferredAlias != null && !preferredAlias.isEmpty()) {
            if (keyStore.containsAlias(preferredAlias)) {
                return preferredAlias;
            }
            logger.warn("Preferred alias '{}' not found in keystore", preferredAlias);
        }

        // Get first available alias
        Enumeration<String> aliases = keyStore.aliases();
        if (aliases.hasMoreElements()) {
            String alias = aliases.nextElement();
            logger.info("Using keystore alias: {}", alias);
            return alias;
        }

        throw new XmlSignerException("ALIAS_ERROR", "No aliases found in keystore", null);
    }

    /**
     * Parse XML string to Document
     */
    private Document parseXmlDocument(String xmlContent) throws Exception {
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        dbf.setNamespaceAware(true);
        dbf.setFeature("http://apache.org/xml/features/disallow-doctype-decl", false);
        dbf.setFeature("http://xml.org/sax/features/external-general-entities", false);
        dbf.setFeature("http://xml.org/sax/features/external-parameter-entities", false);

        DocumentBuilder builder = dbf.newDocumentBuilder();
        InputSource is = new InputSource(new StringReader(xmlContent));
        return builder.parse(is);
    }

    /**
     * Sign XML document
     */
    private void signDocument(Document doc, PrivateKey privateKey,
                              X509Certificate certificate, SignRequest request) throws Exception {

        Element root = doc.getDocumentElement();

        // Determine signature algorithm
        String signatureMethod = getSignatureMethod(request.getSignatureMethod());

        // Create XML Signature
        XMLSignature signature = new XMLSignature(doc, "",
                signatureMethod,
                request.getCanonicalizationMethod());

        root.appendChild(signature.getElement());

        // Add transforms
        Transforms transforms = new Transforms(doc);
        transforms.addTransform(Transforms.TRANSFORM_ENVELOPED_SIGNATURE);
        transforms.addTransform(request.getCanonicalizationMethod());

        // Add document reference
        signature.addDocument("", transforms, Constants.ALGO_ID_DIGEST_SHA256);

        // Add certificate to KeyInfo
        signature.addKeyInfo(certificate);

        // Sign the document
        signature.sign(privateKey);

        logger.debug("Document signed with method: {}", signatureMethod);
    }

    /**
     * Get signature method URI from name
     */
    private String getSignatureMethod(String method) {
        if (method == null || method.isEmpty()) {
            return XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA256;
        }

        switch (method.toUpperCase()) {
            case "RSA-SHA256":
                return XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA256;
            case "RSA-SHA512":
                return XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA512;
            case "RSA-SHA1":
                return XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA1;
            default:
                logger.warn("Unknown signature method '{}', using RSA-SHA256", method);
                return XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA256;
        }
    }

    /**
     * Convert Document to String
     */
    private String documentToString(Document doc) throws Exception {
        TransformerFactory tf = TransformerFactory.newInstance();
        Transformer transformer = tf.newTransformer();
        StringWriter writer = new StringWriter();
        transformer.transform(new DOMSource(doc), new StreamResult(writer));
        return writer.getBuffer().toString();
    }

    /**
     * Get certificate information for logging
     */
    private String getCertificateInfo(X509Certificate certificate) {
        return String.format("Subject: %s, Issuer: %s, Valid from: %s to: %s",
                certificate.getSubjectX500Principal().getName(),
                certificate.getIssuerX500Principal().getName(),
                certificate.getNotBefore(),
                certificate.getNotAfter());
    }

    /**
     * Verify XML signature (for testing)
     */
    public boolean verifySignature(String signedXml) {
        try {
            Document doc = parseXmlDocument(signedXml);
            Element signatureElement = (Element) doc.getElementsByTagNameNS(
                    Constants.SignatureSpecNS, "Signature").item(0);

            if (signatureElement == null) {
                logger.warn("No signature element found in XML");
                return false;
            }

            XMLSignature signature = new XMLSignature(signatureElement, "");
            return signature.checkSignatureValue(signature.getKeyInfo().getPublicKey());

        } catch (Exception e) {
            logger.error("Signature verification failed", e);
            return false;
        }
    }
}
