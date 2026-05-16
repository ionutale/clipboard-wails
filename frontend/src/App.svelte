<script>
  import { onMount } from 'svelte';
  import { GetHistory, CopyToClipboard, DeleteItem, ClearHistory, GetImageData } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';

  let items = [];
  let search = '';
  let toast = '';
  let toastTimer = null;
  let imageCache = {}; // path → data-URL

  // ── load history ──────────────────────────────────────────────────────────
  async function load() {
    try {
      items = (await GetHistory()) || [];
      prefetchImages(items);
    } catch (e) {
      console.error(e);
    }
  }

  async function prefetchImages(list) {
    for (const item of list) {
      if (item.type === 'image' && !imageCache[item.content]) {
        try {
          imageCache[item.content] = await GetImageData(item.content);
        } catch (_) {}
      }
    }
    imageCache = { ...imageCache }; // trigger reactivity
  }

  // ── actions ───────────────────────────────────────────────────────────────
  async function copy(item) {
    await CopyToClipboard(item);
    showToast('Copied!');
  }

  async function remove(item) {
    await DeleteItem(item.id);
    items = items.filter(i => i.id !== item.id);
  }

  async function clearAll() {
    if (!confirm('Clear all clipboard history?')) return;
    await ClearHistory();
    items = [];
    imageCache = {};
  }

  function showToast(msg) {
    toast = msg;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => (toast = ''), 1800);
  }

  // ── time helper ───────────────────────────────────────────────────────────
  function timeAgo(ts) {
    const diff = Date.now() - new Date(ts).getTime();
    const s = Math.floor(diff / 1000);
    if (s < 60)  return `${s}s ago`;
    const m = Math.floor(s / 60);
    if (m < 60)  return `${m}m ago`;
    const h = Math.floor(m / 60);
    if (h < 24)  return `${h}h ago`;
    return `${Math.floor(h / 24)}d ago`;
  }

  // ── realtime ──────────────────────────────────────────────────────────────
  onMount(() => {
    load();

    EventsOn('clipboard-new-item', async (item) => {
      items = [item, ...items.filter(i => i.content !== item.content)];
      if (item.type === 'image' && !imageCache[item.content]) {
        try {
          imageCache[item.content] = await GetImageData(item.content);
          imageCache = { ...imageCache };
        } catch (_) {}
      }
    });

    EventsOn('clipboard-cleared', () => {
      items = [];
      imageCache = {};
    });
  });

  // ── filtered list ─────────────────────────────────────────────────────────
  $: filtered = search
    ? items.filter(i =>
        i.type !== 'image' &&
        i.content.toLowerCase().includes(search.toLowerCase()))
    : items;
</script>

<div class="app">
  <!-- Header -->
  <header>
    <span class="brand">📋 Clipboard</span>
    <input
      class="search"
      type="text"
      placeholder="Search…"
      bind:value={search}
    />
    <button class="clear-btn" on:click={clearAll} title="Clear all history">
      🗑
    </button>
  </header>

  <!-- Items -->
  <main class="list">
    {#if filtered.length === 0}
      <div class="empty">
        {search ? 'No matches.' : 'Nothing copied yet.'}
      </div>
    {/if}

    {#each filtered as item (item.id)}
      <div class="card" on:click={() => copy(item)}>
        {#if item.type === 'image'}
          <div class="card-image">
            {#if imageCache[item.content]}
              <img src={imageCache[item.content]} alt="clipboard image" />
            {:else}
              <span class="img-placeholder">🖼️</span>
            {/if}
          </div>
        {:else}
          <p class="card-text">{item.content}</p>
        {/if}

        <div class="card-meta">
          <span class="badge badge-{item.type}">{item.type}</span>
          <span class="ts">{timeAgo(item.timestamp)}</span>
          <button
            class="del-btn"
            on:click|stopPropagation={() => remove(item)}
            title="Delete"
          >✕</button>
        </div>
      </div>
    {/each}
  </main>

  <!-- Toast -->
  {#if toast}
    <div class="toast">{toast}</div>
  {/if}
</div>

<style>
  :global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }
  :global(body) {
    background: #12121a;
    color: #e0e0e0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 13px;
    height: 100vh;
    overflow: hidden;
    -webkit-app-region: no-drag;
    -webkit-user-select: none;
  }

  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  /* ── header ── */
  header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: #1a1a26;
    border-bottom: 1px solid #2a2a3a;
    flex-shrink: 0;
    -webkit-app-region: drag;
  }

  .brand {
    font-size: 15px;
    font-weight: 600;
    white-space: nowrap;
    -webkit-app-region: no-drag;
  }

  .search {
    flex: 1;
    height: 28px;
    padding: 0 10px;
    background: #24243a;
    border: 1px solid #33334d;
    border-radius: 6px;
    color: #e0e0e0;
    outline: none;
    font-size: 12px;
    -webkit-app-region: no-drag;
  }
  .search::placeholder { color: #555; }
  .search:focus { border-color: #6c6cff; }

  .clear-btn {
    background: none;
    border: none;
    font-size: 16px;
    cursor: pointer;
    opacity: 0.5;
    padding: 2px 4px;
    border-radius: 4px;
    -webkit-app-region: no-drag;
  }
  .clear-btn:hover { opacity: 1; background: #2e2e44; }

  /* ── list ── */
  .list {
    flex: 1;
    overflow-y: auto;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .list::-webkit-scrollbar { width: 5px; }
  .list::-webkit-scrollbar-track { background: transparent; }
  .list::-webkit-scrollbar-thumb { background: #333; border-radius: 4px; }

  /* ── card ── */
  .card {
    background: #1e1e2e;
    border: 1px solid #2a2a3a;
    border-radius: 8px;
    padding: 10px 12px;
    cursor: pointer;
    transition: border-color 0.15s, background 0.15s;
    position: relative;
  }
  .card:hover {
    border-color: #6c6cff;
    background: #22223a;
  }

  .card-text {
    color: #d0d0e0;
    line-height: 1.45;
    max-height: 56px;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 12.5px;
    margin-bottom: 6px;
  }

  .card-image {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    margin-bottom: 6px;
  }
  .card-image img {
    max-height: 90px;
    max-width: 100%;
    border-radius: 4px;
    object-fit: contain;
  }
  .img-placeholder {
    font-size: 36px;
    opacity: 0.4;
  }

  /* ── meta row ── */
  .card-meta {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .badge {
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.4px;
  }
  .badge-text  { background: #2a4a8a; color: #8ab4ff; }
  .badge-image { background: #2a5a3a; color: #7defa6; }
  .badge-file  { background: #4a3a1a; color: #ffd080; }

  .ts {
    color: #555;
    font-size: 10px;
    margin-left: auto;
  }

  .del-btn {
    background: none;
    border: none;
    color: #444;
    cursor: pointer;
    font-size: 11px;
    padding: 2px 5px;
    border-radius: 4px;
    line-height: 1;
    transition: color 0.15s, background 0.15s;
    flex-shrink: 0;
  }
  .del-btn:hover { color: #ff5f5f; background: #2e1a1a; }

  /* ── empty state ── */
  .empty {
    color: #444;
    text-align: center;
    margin-top: 60px;
    font-size: 14px;
  }

  /* ── toast ── */
  .toast {
    position: fixed;
    bottom: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: #6c6cff;
    color: #fff;
    padding: 7px 18px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
    pointer-events: none;
    animation: fadeInUp 0.2s ease;
  }

  @keyframes fadeInUp {
    from { opacity: 0; transform: translateX(-50%) translateY(8px); }
    to   { opacity: 1; transform: translateX(-50%) translateY(0); }
  }
</style>

