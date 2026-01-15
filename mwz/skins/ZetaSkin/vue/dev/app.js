// /app/mwz/skins/ZetaSkin/vue/dev/app.js

let v = localStorage.getItem('v')
if (!v) {
  v = String(Date.now())
  localStorage.setItem('v', v)
}

import(`/config.js?v=${v}`)
  .then(() => import(`/dev5174/src/app.ts?v=${v}`))
  .then(() => console.log('app loaded'))
  .catch(err => {
    console.error('[bootstrap] failed', err)
  })
