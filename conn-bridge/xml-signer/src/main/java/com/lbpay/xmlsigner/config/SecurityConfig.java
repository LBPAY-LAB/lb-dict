package com.lbpay.xmlsigner.config;

import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.web.SecurityFilterChain;

/**
 * Security configuration for XML Signer service
 */
@Configuration
@EnableWebSecurity
public class SecurityConfig {

    /**
     * Configure HTTP Security
     * In dev mode, allow all requests
     * In production, add proper authentication
     */
    @Bean
    @ConditionalOnProperty(name = "app.dev-mode", havingValue = "true", matchIfMissing = true)
    public SecurityFilterChain devSecurityFilterChain(HttpSecurity http) throws Exception {
        http
                .csrf(csrf -> csrf.disable())
                .authorizeHttpRequests(auth -> auth
                        .anyRequest().permitAll()
                );

        return http.build();
    }

    @Bean
    @ConditionalOnProperty(name = "app.dev-mode", havingValue = "false")
    public SecurityFilterChain prodSecurityFilterChain(HttpSecurity http) throws Exception {
        http
                .csrf(csrf -> csrf.disable())
                .authorizeHttpRequests(auth -> auth
                        .requestMatchers("/api/v1/xml-signer/health", "/api/v1/xml-signer/info").permitAll()
                        .anyRequest().authenticated()
                );

        return http.build();
    }
}
