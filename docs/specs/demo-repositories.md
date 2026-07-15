# Substrate Demo Repositories

This document outlines the recommended open-source repositories to fork and use for Substrate demonstrations, stress-testing, and public V2.0 launch materials.

## 1. The "Enterprise Scale" Demo: Stripe API
*   **Repository:** [stripe/openapi](https://github.com/stripe/openapi)
*   **Purpose:** Extreme stress-testing and "shock factor" for the Svelte Flow UI.
*   **Why it works:** Stripe maintains one of the largest and most complex OpenAPI specifications in existence. 
*   **Demo Script:** 
    1. Import the Stripe repo into Substrate.
    2. Show the massive rendering capabilities of the WASM diff engine.
    3. Intentionally remove a deeply nested, critical field (e.g., `charge.amount`).
    4. Demonstrate how Substrate catches the breaking change instantly in the Visual API Studio.

## 2. The "Cross-Service Blast Radius" Demo: Google Cloud Microservices
*   **Repository:** [GoogleCloudPlatform/microservices-demo](https://github.com/GoogleCloudPlatform/microservices-demo)
*   **Purpose:** Demonstrating the true value of Substrate's cross-repo dependency graph.
*   **Why it works:** This is a realistic 10-tier e-commerce microservices application (Checkout, Payment, Email, Currency, etc.) communicating via gRPC/Protobufs and REST.
*   **Demo Script:**
    1. Map the individual microservice folders as separate consumers/providers in `substrate.yaml`.
    2. Open the Substrate Dashboard to reveal the complex 10-node dependency graph.
    3. Introduce a breaking change in the `payment-service`.
    4. Watch the UI's "Blast Radius" feature dynamically highlight the `checkout-service` in red, proving that Substrate detects downstream cascading failures.

## 3. The "Standard SaaS" Demo: RealWorld (Conduit)
*   **Repository:** [gothinkster/realworld](https://github.com/gothinkster/realworld)
*   **Purpose:** The standard "Zero-to-One" onboarding demo for new users.
*   **Why it works:** RealWorld is the modern "TodoMVC". It has dozens of backend implementations that all strictly adhere to a single standardized OpenAPI spec for a Medium.com clone.
*   **Demo Script:**
    1. Go through the Substrate Onboarding Wizard using a RealWorld fork.
    2. Drop in a `substrate.yaml` configuration.
    3. Open a GitHub Pull Request that changes the `/api/articles` endpoint response format.
    4. Show the Substrate GitHub App automatically commenting on the PR and blocking the merge.
