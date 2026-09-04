<script lang="ts">
  import { onMount } from 'svelte';

  let isModalOpen = $state(false);
  let attributionsContent = $state('Loading licenses...');

  function toggleModal() {
    isModalOpen = !isModalOpen;
  }

  onMount(async () => {
    try {
      const response = await fetch('/ATTRIBUTIONS.md');
      if (response.ok) {
        attributionsContent = await response.text();
      } else {
        attributionsContent = 'Could not load attributions content directly. Please see ATTRIBUTIONS.md in the repository root.';
      }
    } catch (e) {
      attributionsContent = 'Could not load attributions content directly. Please see ATTRIBUTIONS.md in the repository root.';
    }
  });
</script>

<footer class="mt-8 py-4 text-center text-sm text-gray-500">
  <button type="button" class="underline hover:text-gray-700 cursor-pointer" onclick={toggleModal}>
    Credits / OSS Licenses
  </button>
</footer>

{#if isModalOpen}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white p-6 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] flex flex-col">
      <div class="flex justify-between items-center mb-4">
        <h2 class="text-xl font-bold">Credits / OSS Licenses</h2>
        <button type="button" class="text-gray-500 hover:text-gray-700 text-2xl font-bold cursor-pointer" onclick={toggleModal}>
          &times;
        </button>
      </div>

      <div class="overflow-y-auto flex-grow p-4 border rounded bg-gray-50">
        <p class="mb-4">Substrate is made possible by the following open-source software:</p>
        <div class="whitespace-pre-wrap font-mono text-xs text-left overflow-x-auto">
          {attributionsContent}
        </div>
      </div>

      <div class="mt-6 flex justify-end">
        <button type="button" class="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded cursor-pointer" onclick={toggleModal}>
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
