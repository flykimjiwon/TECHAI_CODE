// 택가이 컴패니언 — SSE client + DOM manager
(function () {
  'use strict';

  // --- DOM refs ---
  const chatBody = document.getElementById('chat-body');
  const planBody = document.getElementById('plan-body');
  const toolsBody = document.getElementById('tools-body');
  const statusDot = document.getElementById('status-indicator');
  const modeBadge = document.getElementById('mode-badge');
  const autoBadge = document.getElementById('auto-badge');
  const sessionBadge = document.getElementById('session-badge');
  const toolCount = document.getElementById('tool-count');

  const MODE_NAMES = ['Code', 'Deep', 'Plan'];
  const MODE_CLASSES = ['mode-0', 'mode-1', 'mode-2'];

  let lastEventId = 0;
  let toolEntryCount = 0;
  let streamingEl = null; // current streaming message element
  let reconnectDelay = 1000;
  let mermaidCounter = 0;

  // --- SSE Connection ---
  function connect() {
    const url = lastEventId ? '/events?after=' + lastEventId : '/events';
    const es = new EventSource(url);

    es.onopen = function () {
      setStatus('connected');
      reconnectDelay = 1000;
    };

    es.onerror = function () {
      es.close();
      setStatus('reconnecting');
      setTimeout(connect, reconnectDelay);
      reconnectDelay = Math.min(reconnectDelay * 2, 30000);
    };

    // Each named event type gets its own listener.
    var types = [
      'state_snapshot', 'user_message', 'stream_start', 'stream_chunk',
      'stream_done', 'tool_call_start', 'tool_result', 'mode_change',
      'auto_toggle', 'multi_progress', 'multi_result',
      'session_create', 'session_load'
    ];
    types.forEach(function (t) {
      es.addEventListener(t, function (e) {
        var data;
        try { data = JSON.parse(e.data); } catch (_) { return; }
        if (data.id && data.id > lastEventId) lastEventId = data.id;
        handleEvent(data.type || t, data.data || data);
      });
    });
  }

  function setStatus(state) {
    statusDot.className = 'status-dot ' + state;
    statusDot.title = state === 'connected' ? '연결됨' :
      state === 'reconnecting' ? '재연결 중...' : '연결 끊김';
  }

  // --- Event Dispatch ---
  function handleEvent(type, data) {
    switch (type) {
      case 'state_snapshot': onStateSnapshot(data); break;
      case 'user_message': onUserMessage(data); break;
      case 'stream_start': onStreamStart(data); break;
      case 'stream_chunk': onStreamChunk(data); break;
      case 'stream_done': onStreamDone(data); break;
      case 'tool_call_start': onToolCallStart(data); break;
      case 'tool_result': onToolResult(data); break;
      case 'mode_change': onModeChange(data); break;
      case 'auto_toggle': onAutoToggle(data); break;
      case 'multi_progress': onMultiProgress(data); break;
      case 'multi_result': onMultiResult(data); break;
      case 'session_create': onSessionCreate(data); break;
      case 'session_load': onSessionLoad(data); break;
    }
  }

  // --- Event Handlers ---

  function onStateSnapshot(d) {
    if (d.mode !== undefined) setMode(d.mode);
    if (d.autoMode) autoBadge.classList.remove('hidden');
    if (d.sessionID) sessionBadge.textContent = '#' + d.sessionID;
    addSystemMessage('컴패니언 연결됨 (메시지 ' + (d.messages || 0) + '개)');
  }

  function onUserMessage(d) {
    addMessage('user', d.content || '');
  }

  function onStreamStart(_d) {
    // Create a streaming placeholder in the plan panel
    streamingEl = document.createElement('div');
    streamingEl.className = 'message assistant streaming';
    streamingEl.innerHTML = '<span class="role-tag">assistant</span><span class="stream-content"></span><span class="streaming-cursor"></span>';
    planBody.appendChild(streamingEl);
    autoScroll(planBody);
  }

  function onStreamChunk(d) {
    if (!streamingEl) return;
    var content = streamingEl.querySelector('.stream-content');
    if (content) {
      content.textContent += (d.content || '');
      autoScroll(planBody);
    }
  }

  function onStreamDone(d) {
    // Finalize the streaming element
    if (streamingEl) {
      var cursor = streamingEl.querySelector('.streaming-cursor');
      if (cursor) cursor.remove();
      streamingEl.classList.remove('streaming');

      // Render markdown in the plan panel
      var content = streamingEl.querySelector('.stream-content');
      if (content) {
        content.innerHTML = renderMarkdown(content.textContent);
        highlightCode(content);
        renderMermaid(content);
      }
      streamingEl = null;
    }

    // Also add final message to chat panel
    var text = d.content || '';
    if (text) {
      var el = addMessage('assistant', text);
      highlightCode(el);
      renderMermaid(el);
    }

    // Show stats
    if (d.tokens || d.elapsed) {
      var stats = '';
      if (d.tokens) stats += d.tokens + 'tok';
      if (d.elapsed) stats += (stats ? ' | ' : '') + d.elapsed.toFixed(1) + 's';
      addSystemMessage(stats);
    }
  }

  function onToolCallStart(d) {
    toolEntryCount++;
    toolCount.textContent = toolEntryCount;

    var entry = document.createElement('div');
    entry.className = 'tool-entry';
    entry.id = 'tool-' + (d.id || toolEntryCount);
    entry.innerHTML =
      '<div class="tool-entry-header">' +
        '<span class="tool-status running">&#9679;</span>' +
        '<span class="tool-name">' + esc(d.name || '') + '</span>' +
        '<span class="tool-args">' + esc(d.args || '') + '</span>' +
      '</div>' +
      '<div class="tool-entry-body"></div>';

    entry.querySelector('.tool-entry-header').onclick = function () {
      entry.classList.toggle('expanded');
    };

    toolsBody.appendChild(entry);
    autoScroll(toolsBody);

    // Also note in chat
    addMessage('tool', '>> ' + (d.name || '') + '(' + truncate(d.args || '', 80) + ')');
  }

  function onToolResult(d) {
    // Update the last running tool entry
    var entries = toolsBody.querySelectorAll('.tool-entry');
    for (var i = entries.length - 1; i >= 0; i--) {
      var status = entries[i].querySelector('.tool-status');
      if (status && status.classList.contains('running')) {
        status.classList.remove('running');
        status.classList.add('done');
        status.innerHTML = '&#10003;';
        var body = entries[i].querySelector('.tool-entry-body');
        if (body) body.textContent = d.output || '';
        break;
      }
    }

    addMessage('tool', '<< ' + (d.name || '') + ': ' + truncate(d.output || '', 120));
  }

  function onModeChange(d) {
    if (d.mode !== undefined) setMode(d.mode);
  }

  function onAutoToggle(d) {
    if (d.enabled) {
      autoBadge.classList.remove('hidden');
    } else {
      autoBadge.classList.add('hidden');
    }
  }

  function onMultiProgress(d) {
    var id = 'multi-' + (d.agent || 'unknown');
    var el = document.getElementById(id);
    if (!el) {
      el = document.createElement('div');
      el.className = 'multi-progress';
      el.id = id;
      planBody.appendChild(el);
    }
    el.innerHTML =
      '<span class="multi-agent-name">' + esc(d.agent || '') + '</span> ' +
      '<span class="multi-status">' + esc(d.status || '') + '</span> ' +
      '<span style="color:var(--text-dim)">' + esc(d.detail || '') + '</span>';
    autoScroll(planBody);
  }

  function onMultiResult(d) {
    // Remove progress elements
    var progs = planBody.querySelectorAll('.multi-progress');
    progs.forEach(function (el) { el.remove(); });

    if (d.output) {
      var el = addMessage('assistant', d.output);
      highlightCode(el);
    }
    var stats = (d.strategy || '') + ' | ' + (d.tokens || 0) + 'tok';
    if (d.elapsed) stats += ' | ' + d.elapsed.toFixed(1) + 's';
    addSystemMessage('[Multi] ' + stats);
  }

  function onSessionCreate(d) {
    sessionBadge.textContent = '#' + (d.id || '');
    addSystemMessage('새 세션 #' + (d.id || ''));
  }

  function onSessionLoad(d) {
    sessionBadge.textContent = '#' + (d.id || '');
    chatBody.innerHTML = '';
    planBody.innerHTML = '';
    toolsBody.innerHTML = '';
    toolEntryCount = 0;
    toolCount.textContent = '0';
    addSystemMessage('세션 #' + (d.id || '') + " '" + (d.title || '') + "' 복원 (" + (d.messages || 0) + '개)');
  }

  // --- DOM helpers ---

  function addMessage(role, text) {
    var el = document.createElement('div');
    el.className = 'message ' + role;

    var tag = '';
    if (role === 'user') tag = 'user';
    else if (role === 'assistant') tag = 'assistant';
    else if (role === 'system') tag = 'system';

    if (tag) {
      el.innerHTML = '<span class="role-tag">' + tag + '</span>' + renderMarkdown(text);
    } else {
      el.textContent = text;
    }

    chatBody.appendChild(el);
    autoScroll(chatBody);
    return el;
  }

  function addSystemMessage(text) {
    return addMessage('system', text);
  }

  function setMode(mode) {
    modeBadge.textContent = MODE_NAMES[mode] || 'Code';
    modeBadge.className = 'mode-badge ' + (MODE_CLASSES[mode] || 'mode-0');
  }

  function autoScroll(container) {
    var threshold = 100;
    var atBottom = container.scrollHeight - container.scrollTop - container.clientHeight < threshold;
    if (atBottom) {
      requestAnimationFrame(function () {
        container.scrollTop = container.scrollHeight;
      });
    }
  }

  // --- Minimal Markdown ---

  function renderMarkdown(text) {
    if (!text) return '';
    var html = esc(text);

    // Code blocks (```lang\n...\n```)
    html = html.replace(/```(\w*)\n([\s\S]*?)```/g, function (_, lang, code) {
      var cls = lang ? ' class="language-' + lang + '"' : '';
      return '<pre><code' + cls + '>' + code + '</code></pre>';
    });

    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

    // Headers
    html = html.replace(/^#### (.+)$/gm, '<h4>$1</h4>');
    html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
    html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
    html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

    // Bold and italic
    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    html = html.replace(/\*([^*]+)\*/g, '<em>$1</em>');

    // Unordered lists
    html = html.replace(/^- (.+)$/gm, '<li>$1</li>');

    // Ordered lists
    html = html.replace(/^\d+\. (.+)$/gm, '<li>$1</li>');

    // Links
    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>');

    // Line breaks
    html = html.replace(/\n/g, '<br>');

    return html;
  }

  function esc(s) {
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(s));
    return d.innerHTML;
  }

  function truncate(s, max) {
    if (s.length <= max) return s;
    return s.substring(0, max) + '...';
  }

  // --- Code highlighting (optional) ---

  function highlightCode(container) {
    if (!container || !window.hljs) return;
    var blocks = container.querySelectorAll('pre code');
    blocks.forEach(function (block) {
      window.hljs.highlightElement(block);
    });
  }

  // --- Mermaid rendering (optional) ---

  function renderMermaid(container) {
    if (!container || !window.mermaid) return;
    var blocks = container.querySelectorAll('pre code.language-mermaid');
    blocks.forEach(function (block) {
      var pre = block.parentElement;
      var wrapper = document.createElement('div');
      wrapper.className = 'mermaid-container';
      var mermaidDiv = document.createElement('div');
      mermaidDiv.className = 'mermaid';
      mermaidDiv.id = 'mermaid-' + (++mermaidCounter);
      mermaidDiv.textContent = block.textContent;
      wrapper.appendChild(mermaidDiv);
      pre.replaceWith(wrapper);

      try {
        window.mermaid.init(undefined, mermaidDiv);
      } catch (e) {
        mermaidDiv.textContent = 'Mermaid error: ' + e.message;
      }
    });
  }

  // --- Keyboard shortcuts ---

  document.addEventListener('keydown', function (e) {
    // Escape: collapse all tool entries
    if (e.key === 'Escape') {
      document.querySelectorAll('.tool-entry.expanded').forEach(function (el) {
        el.classList.remove('expanded');
      });
    }
    // Ctrl+L: clear chat display
    if (e.ctrlKey && e.key === 'l') {
      e.preventDefault();
      chatBody.innerHTML = '';
      planBody.innerHTML = '';
    }
  });

  // --- Init ---

  if (window.mermaid) {
    window.mermaid.initialize({ theme: 'dark', startOnLoad: false });
  }

  setStatus('disconnected');
  connect();
})();
