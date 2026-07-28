# P12-T09: AI Support Copilot

## Objective
Implement a floating AI chat widget in the Substrate Dashboard that acts as an intelligent, real-time assistant for users. The copilot should leverage the existing Phase 4 Intelligence Layer to answer Substrate-specific questions, help write valid `substrate.yaml` files, and troubleshoot graph validation errors.

## Context
Substrate provides a rich visual Dependency Graph and a Visual Studio for modifying `substrate.yaml` contracts. However, users occasionally struggle with syntax, identifying why a breaking change occurred, or how to define complex cross-repo dependencies. This widget will serve as an ever-present guide, seamlessly tying into the AI capabilities already established in the Go API backend.

## Requirements

### 1. UI/UX (SvelteKit)
- **Floating Widget:** A sticky, collapsible chat bubble fixed to the bottom-right corner of the screen (`fixed bottom-4 right-4`).
- **Premium Aesthetics:** Align with Substrate's design system—use glassmorphism, smooth CSS transitions for opening/closing, and the standard dark mode color palette.
- **Chat Interface:**
  - A scrollable message history container.
  - Distinct styling for user messages vs. AI responses (e.g., Markdown rendering for AI responses to handle code blocks).
  - An input field with a send button (supporting `Enter` to send).
  - A "Typing..." indicator or streaming text effect while waiting for the AI.
- **Context Awareness:** The widget should be able to read the current state (e.g., the currently selected Node in the Graph, or the text in the AI Playground editor) to inject into the AI prompt as context.

### 2. Backend Integration
- The frontend should communicate with the existing Phase 4 Intelligence Layer endpoint (`/api/v1/ai/chat` or similar).
- If an endpoint does not exist specifically for chat, design the frontend service to use Server-Sent Events (SSE) against the existing `/api/v1/ai/analyze` by constructing a prompt that asks the AI for chat assistance.
- **Mock Fallback:** Ensure that during development or E2E tests, the backend call is gracefully mocked out.

### 3. State Management
- Use Svelte 5 Runes (`$state`) to manage the array of messages (`{ role: 'user' | 'ai', content: string }[]`).
- Persist the chat history across dashboard route changes (e.g., using SvelteKit stores or local storage) so the user doesn't lose their conversation if they navigate from the Dashboard to the Studio.

## Acceptance Criteria
1. The AI widget is visible on all authenticated dashboard routes and can be toggled open and closed smoothly.
2. A user can type a question and receive a Markdown-formatted response from the real `/api/v1/ai/analyze` endpoint (or its deterministic fallback when no LLM is configured).
3. The component handles network loading states and potential API errors gracefully.
4. E2E tests verify the widget opens, accepts text input, and renders the real API response. **No `page.route()` mocking.**
