package com.lbpay.xmlsigner;

import com.lbpay.xmlsigner.model.SignRequest;
import com.lbpay.xmlsigner.model.SignResponse;
import com.lbpay.xmlsigner.service.XmlSignerService;
import com.lbpay.xmlsigner.util.CertificateGenerator;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.io.File;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Integration tests for XML Signer Service
 */
@SpringBootTest
public class XmlSignerServiceTest {

    @Autowired
    private XmlSignerService xmlSignerService;

    private static final String TEST_CERT_PATH = "target/test-cert.p12";
    private static final String TEST_CERT_PASSWORD = "testpass";
    private static final String TEST_ALIAS = "test";

    @BeforeAll
    public static void setUp() throws Exception {
        // Generate test certificate
        File certFile = new File(TEST_CERT_PATH);
        if (!certFile.exists()) {
            CertificateGenerator.generateSelfSignedCertificate(
                    TEST_CERT_PATH,
                    TEST_CERT_PASSWORD,
                    TEST_ALIAS,
                    "Test Certificate",
                    365
            );
        }
    }

    @Test
    public void testSignXml_Success() {
        // Arrange
        String xmlContent = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
                "<root>\n" +
                "  <data>Test content</data>\n" +
                "</root>";

        SignRequest request = new SignRequest();
        request.setXmlContent(xmlContent);
        request.setCertificatePath(TEST_CERT_PATH);
        request.setCertificatePassword(TEST_CERT_PASSWORD);
        request.setDevMode(true);
        request.setKeyAlias(TEST_ALIAS);

        // Act
        SignResponse response = xmlSignerService.signXml(request);

        // Assert
        assertTrue(response.isSuccess(), "Signing should succeed");
        assertNotNull(response.getSignedXml(), "Signed XML should not be null");
        assertTrue(response.getSignedXml().contains("Signature"), "Signed XML should contain Signature element");
        assertNotNull(response.getCertificateInfo(), "Certificate info should be present");
        assertNull(response.getError(), "Error should be null on success");
    }

    @Test
    public void testSignAndVerify() {
        // Arrange
        String xmlContent = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
                "<document>\n" +
                "  <content>Verification test</content>\n" +
                "</document>";

        SignRequest request = new SignRequest();
        request.setXmlContent(xmlContent);
        request.setCertificatePath(TEST_CERT_PATH);
        request.setCertificatePassword(TEST_CERT_PASSWORD);
        request.setDevMode(true);

        // Act - Sign
        SignResponse signResponse = xmlSignerService.signXml(request);

        // Assert - Sign
        assertTrue(signResponse.isSuccess(), "Signing should succeed");

        // Act - Verify
        boolean isValid = xmlSignerService.verifySignature(signResponse.getSignedXml());

        // Assert - Verify
        assertTrue(isValid, "Signature should be valid");
    }

    @Test
    public void testSignXml_InvalidCertPath() {
        // Arrange
        SignRequest request = new SignRequest();
        request.setXmlContent("<root><data>test</data></root>");
        request.setCertificatePath("/invalid/path/cert.p12");
        request.setCertificatePassword("password");
        request.setDevMode(true);

        // Act
        SignResponse response = xmlSignerService.signXml(request);

        // Assert
        assertFalse(response.isSuccess(), "Signing should fail with invalid cert path");
        assertNotNull(response.getError(), "Error message should be present");
    }

    @Test
    public void testSignXml_InvalidXml() {
        // Arrange
        SignRequest request = new SignRequest();
        request.setXmlContent("This is not XML");
        request.setCertificatePath(TEST_CERT_PATH);
        request.setCertificatePassword(TEST_CERT_PASSWORD);
        request.setDevMode(true);

        // Act
        SignResponse response = xmlSignerService.signXml(request);

        // Assert
        assertFalse(response.isSuccess(), "Signing should fail with invalid XML");
        assertNotNull(response.getError(), "Error message should be present");
    }

    @Test
    public void testSignXml_DifferentAlgorithms() {
        // Test RSA-SHA256
        testWithAlgorithm("RSA-SHA256");

        // Test RSA-SHA512
        testWithAlgorithm("RSA-SHA512");
    }

    private void testWithAlgorithm(String algorithm) {
        SignRequest request = new SignRequest();
        request.setXmlContent("<root><data>Algorithm test</data></root>");
        request.setCertificatePath(TEST_CERT_PATH);
        request.setCertificatePassword(TEST_CERT_PASSWORD);
        request.setDevMode(true);
        request.setSignatureMethod(algorithm);

        SignResponse response = xmlSignerService.signXml(request);

        assertTrue(response.isSuccess(), "Signing should succeed with " + algorithm);
        assertNotNull(response.getSignedXml(), "Signed XML should not be null");
    }
}
