// buildHtml.ts
export default function buildHtml(
  id: string,
  html: string,
  js: string,
): string {
  const raw = html.trim()
  const hasHtmlTag = /^<html[\s>]/i.test(raw)

  const escapedJs = js
    .replace(/\\/g, '\\\\')
    .replace(/`/g, '\\`')
    .replace(/\$\{/g, '\\${')

  const consoleProxyScript = `
<script>
(function () {
  function send(level, args) {
    var argArray = Array.prototype.slice.call(args);
    try {
      if (window.parent && window.parent !== window) {
        var fn = window.parent['${id}'];
        if (typeof fn === 'function') {
          fn({ level: level, args: argArray });
        }
      }
    } catch (e) {
      // ignore
    }
  }

  // console proxy
  var proxy = {};
  ['log','info','warn','error','debug','trace'].forEach(function (level) {
    proxy[level] = function () {
      send(level, arguments);
    };
  });
  window.console = proxy;

  window.addEventListener('error', function (event) {
    try {
      var msg;
      if (event.error && event.error.name && event.error.message) {
        msg = event.error.name + ': ' + event.error.message;
      } else {
        msg = event.message || 'Unknown script error';
      }
      var loc = '';
      if (event.filename) {
        loc += ' (' + event.filename;
        if (event.lineno) {
          loc += ':' + event.lineno;
          if (event.colno) loc += ':' + event.colno;
        }
        loc += ')';
      }
      send('error', [ 'Uncaught ' + msg + loc ]);
    } catch (e) {}
  });

  window.addEventListener('unhandledrejection', function (event) {
    try {
      var r = event.reason;
      var msg;
      if (r && r.name && r.message) {
        msg = r.name + ': ' + r.message;
      } else {
        msg = String(r);
      }
      send('error', [ 'Unhandled promise rejection: ' + msg ]);
    } catch (e) {}
  });
})();
<\/script>
`

  const execScript = `
<script>
(function () {
  try {
    new Function(\`${escapedJs}\`)();
  } catch (e) {
    console.error('Uncaught ' + e.name + ': ' + e.message);
  }
})();
<\/script>
`

  let baseHtml: string
  if (hasHtmlTag) {
    baseHtml = raw
  } else {
    baseHtml = `<html><body>${raw}</body></html>`
  }

  let htmlWithProxy = baseHtml

  if (/<head>/i.test(htmlWithProxy)) {
    htmlWithProxy = htmlWithProxy.replace(
      /<head>/i,
      `<head>${consoleProxyScript}`,
    )
  } else if (/<body[^>]*>/i.test(htmlWithProxy)) {
    htmlWithProxy = htmlWithProxy.replace(
      /<body[^>]*>/i,
      `$&${consoleProxyScript}`,
    )
  } else {
    htmlWithProxy = consoleProxyScript + htmlWithProxy
  }

  if (/<\/body>/i.test(htmlWithProxy)) {
    return htmlWithProxy.replace(/<\/body>/i, `${execScript}\n</body>`)
  }
  if (/<\/html>/i.test(htmlWithProxy)) {
    return htmlWithProxy.replace(/<\/html>/i, `${execScript}\n</html>`)
  }
  return htmlWithProxy + execScript
}
