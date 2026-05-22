function copyCmd(btn, text) {
  var prev = btn.innerHTML;
  var show = function() {
    btn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><polyline points="20 6 9 17 4 12"/></svg> copied';
    btn.style.color = 'var(--green)';
    setTimeout(function() { btn.innerHTML = prev; btn.style.color = ''; }, 2000);
  };
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(show).catch(function() {
      fallbackCopy(text); show();
    });
  } else {
    fallbackCopy(text); show();
  }
}

function fallbackCopy(text) {
  var el = document.createElement('textarea');
  el.value = text;
  el.style.position = 'fixed';
  el.style.opacity = '0';
  document.body.appendChild(el);
  el.select();
  try { document.execCommand('copy'); } catch(e) {}
  document.body.removeChild(el);
}
