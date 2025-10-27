package com.lbpay.xmlsigner;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * XML Signer Application - Digital Signature Service for ICP-Brasil
 *
 * Main Spring Boot application for signing XML documents with ICP-Brasil A3 certificates.
 * Supports both production (hardware tokens/A3) and development (self-signed) modes.
 */
@SpringBootApplication
public class XmlSignerApplication {

    public static void main(String[] args) {
        // Initialize Apache Santuario XML Security library
        org.apache.xml.security.Init.init();

        SpringApplication.run(XmlSignerApplication.class, args);
    }

    /**
     * Configure CORS for development
     */
    @Bean
    public WebMvcConfigurer corsConfigurer() {
        return new WebMvcConfigurer() {
            @Override
            public void addCorsMappings(CorsRegistry registry) {
                registry.addMapping("/**")
                        .allowedOrigins("*")
                        .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
                        .allowedHeaders("*");
            }
        };
    }
}
