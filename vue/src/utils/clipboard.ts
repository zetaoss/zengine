export default async function copyToClipboard(s: string) {
  try {
    await navigator.clipboard.writeText(s);
    return true;
  } catch (err) {
    console.error("Failed to copy text to clipboard:", err);
    return false;
  }
}
