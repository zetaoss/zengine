// scroll.ts

export function scrollToTop() {
  window.scrollTo({
    top: 0,
    behavior: 'smooth',
  })
}

export function scrollToBottom() {
  window.scrollTo({
    top: document.documentElement.scrollHeight,
    behavior: 'smooth',
  })
}
