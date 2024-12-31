export default function stripTags(s: String) {
  return s.replace(/<\/?[^>]+>/ig, ' ')
}
