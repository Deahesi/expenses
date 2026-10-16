export function colorByTextHash(text: string): string {
  let hash = 0

  for (let index = 0; index < text.length; index++) {
    hash = (Math.imul(hash, 31) + text.charCodeAt(index)) | 0
  }

  return `hsl(${(hash >>> 0) % 360} 60% 30%)`
}
