// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import MessageSquare from '@lucide/svelte/icons/message-square';
	import X from '@lucide/svelte/icons/x';
	import Send from '@lucide/svelte/icons/send';
	import { onMount } from 'svelte';

	interface Message {
		role: 'user' | 'ai';
		content: string;
	}

	let isOpen = $state(false);
	let messages: Message[] = $state([]);
	let inputMessage = $state('');
	let widgetRef: HTMLDivElement;
	let isTyping = $state(false);
	let messagesContainer: HTMLDivElement = $state() as HTMLDivElement;

	function toggleWidget() {
		isOpen = !isOpen;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && isOpen) {
			isOpen = false;
		}
	}

	function handleClickOutside(event: MouseEvent) {
		if (isOpen && widgetRef && !widgetRef.contains(event.target as Node)) {
			// Svelte Flow may intercept clicks, so we only close if the click originated from outside
			// Ensure we are checking if the target is an element, otherwise we might error out
			isOpen = false;
		}
	}

	function preventPropagation(event: MouseEvent) {
		event.stopPropagation();
	}

	onMount(() => {
		// Load from local storage
		const savedMessages = localStorage.getItem('copilot_messages');
		if (savedMessages) {
			try {
				messages = JSON.parse(savedMessages);
			} catch (e) {
				console.error('Failed to parse saved messages');
			}
		}

		window.addEventListener('keydown', handleKeydown);
		window.addEventListener('mousedown', handleClickOutside);

		return () => {
			window.removeEventListener('keydown', handleKeydown);
			window.removeEventListener('mousedown', handleClickOutside);
		};
	});

	$effect(() => {
		if (typeof window !== 'undefined') {
			localStorage.setItem('copilot_messages', JSON.stringify(messages));
		}
	});

	$effect(() => {
		if (messagesContainer && (messages.length || isTyping)) {
			messagesContainer.scrollTop = messagesContainer.scrollHeight;
		}
	});

	async function sendMessage() {
		if (!inputMessage.trim()) return;

		messages = [...messages, { role: 'user', content: inputMessage }];
		let currentMessage = inputMessage;
		inputMessage = '';
		isTyping = true;

		// Add an empty AI message to stream into
		messages = [...messages, { role: 'ai', content: '' }];
		let aiMessageIndex = messages.length - 1;

		try {
			// Construct prompt for AI analyze using the current message history
			// We format history as text and pass it to current_schema to leverage existing endpoint
			let chatHistory = messages.slice(0, -1).map(m => `${m.role === 'user' ? 'User' : 'Assistant'}: ${m.content}`).join('\n');

			const response = await fetch('/api/v1/ai/analyze', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					org: 'default',
					current_schema: chatHistory,
					proposed_schema: currentMessage,
					schema_type: 'chat'
				})
			});

			if (!response.ok) {
				throw new Error(`API error: ${response.status}`);
			}

			const reader = response.body?.getReader();
			const decoder = new TextDecoder();
			let buffer = '';

			if (!reader) {
				throw new Error('No reader available');
			}

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');

				// Keep the last partial line in the buffer
				buffer = lines.pop() || '';

				for (const line of lines) {
					if (line.trim() === '') continue;
					if (line.startsWith('data: ')) {
						const dataStr = line.slice(6);
						if (dataStr === '[DONE]') {
							continue;
						}

						try {
							const event = JSON.parse(dataStr);

							if (event.type === 'thinking' && event.content) {
								messages[aiMessageIndex].content += event.content;
							} else if (event.type === 'finding' && event.content) {
								if (!messages[aiMessageIndex].content.includes(event.content)) {
									messages[aiMessageIndex].content += `\n\n**${event.severity}**: ${event.content}`;
								}
							} else if (event.type === 'fix' && event.code) {
								messages[aiMessageIndex].content += `\n\n\`\`\`${event.language || ''}\n${event.code}\n\`\`\``;
							} else if (event.type === 'error') {
								messages[aiMessageIndex].content += `\n\n*Error: ${event.content}*`;
							}
							// Reassign to trigger Svelte reactivity
							messages = [...messages];
						} catch (e) {
							console.error('Failed to parse SSE message:', e);
						}
					}
				}
			}
		} catch (error) {
			console.error('Chat error:', error);
			messages[aiMessageIndex].content = 'Sorry, I encountered an error communicating with the server.';
			messages = [...messages];
		} finally {
			isTyping = false;
		}
	}

	function handleInputKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			sendMessage();
		}
	}
</script>

<div class="fixed bottom-6 right-6 z-50" bind:this={widgetRef} onmousedown={preventPropagation} role="presentation">
	{#if isOpen}
		<div
			class="mb-4 w-80 sm:w-96 rounded-lg shadow-2xl overflow-hidden flex flex-col border border-[var(--border)] bg-[var(--bg-card)] backdrop-blur-md transition-all duration-300 ease-in-out"
			style="height: 500px; max-height: calc(100vh - 100px);"
		>
			<!-- Header -->
			<div class="bg-opacity-20 bg-[var(--bg-hover)] border-b border-[var(--border)] p-4 flex justify-between items-center">
				<h3 class="font-medium text-[var(--text-main)] flex items-center gap-2">
					<MessageSquare class="w-5 h-5 text-[var(--accent)]" />
					Support Copilot
				</h3>
				<button
					class="text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors p-1 rounded hover:bg-[var(--bg-hover)]"
					onclick={toggleWidget}
					aria-label="Close Copilot"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Messages -->
			<div class="flex-grow p-4 overflow-y-auto flex flex-col gap-4 bg-[var(--bg-dark)] bg-opacity-50" bind:this={messagesContainer}>
				{#if messages.length === 0}
					<div class="text-center text-[var(--text-muted)] mt-10 text-sm">
						Hi! I'm your Substrate Support Copilot. Ask me about your graph, substrate.yaml, or anything else!
					</div>
				{/if}

				{#each messages as msg}
					<div class="flex flex-col {msg.role === 'user' ? 'items-end' : 'items-start'}">
						<div
							class="max-w-[85%] rounded-lg p-3 text-sm {msg.role === 'user'
								? 'bg-[var(--accent)] text-white'
								: 'bg-[var(--bg-card)] border border-[var(--border)] text-[var(--text-main)]'}"
						>
							{#if msg.role === 'ai'}
								<!-- Basic Markdown rendering for AI responses -->
								<div class="whitespace-pre-wrap">{msg.content}</div>
							{:else}
								<div class="whitespace-pre-wrap">{msg.content}</div>
							{/if}
						</div>
					</div>
				{/each}

				{#if isTyping}
					<div class="flex flex-col items-start">
						<div class="max-w-[85%] rounded-lg p-3 text-sm bg-[var(--bg-card)] border border-[var(--border)] text-[var(--text-main)] flex items-center gap-2">
							<div class="w-2 h-2 rounded-full bg-[var(--accent)] animate-bounce" style="animation-delay: 0ms"></div>
							<div class="w-2 h-2 rounded-full bg-[var(--accent)] animate-bounce" style="animation-delay: 150ms"></div>
							<div class="w-2 h-2 rounded-full bg-[var(--accent)] animate-bounce" style="animation-delay: 300ms"></div>
						</div>
					</div>
				{/if}
			</div>

			<!-- Input -->
			<div class="p-3 border-t border-[var(--border)] bg-[var(--bg-card)]">
				<div class="flex items-end gap-2 bg-[var(--bg-dark)] border border-[var(--border)] rounded-lg p-1 focus-within:border-[var(--accent)] transition-colors">
					<textarea
						bind:value={inputMessage}
						onkeydown={handleInputKeydown}
						placeholder="Ask a question..."
						class="flex-grow bg-transparent border-none outline-none text-sm p-2 resize-none text-[var(--text-main)] placeholder-[var(--text-muted)] max-h-32 min-h-[40px]"
						rows="1"
					></textarea>
					<button
						onclick={sendMessage}
						disabled={!inputMessage.trim() || isTyping}
						class="p-2 mb-0.5 mr-0.5 rounded-md text-white bg-[var(--accent)] hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
						aria-label="Send message"
					>
						<Send class="w-4 h-4" />
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Floating Action Button -->
	<button
		class="w-14 h-14 rounded-full bg-[var(--accent)] text-white shadow-lg flex items-center justify-center hover:scale-105 transition-transform duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-[var(--bg-dark)] focus:ring-[var(--accent)] ml-auto"
		onclick={toggleWidget}
		aria-label={isOpen ? "Close Support Copilot" : "Open Support Copilot"}
	>
		{#if isOpen}
			<X class="w-6 h-6" />
		{:else}
			<MessageSquare class="w-6 h-6" />
		{/if}
	</button>
</div>
