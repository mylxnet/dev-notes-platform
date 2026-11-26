// 链接标识（slug）生成：中文转拼音、字符清洗、长度截断与重名加后缀。
// 只在后台「新建」表单使用；编辑已有内容时不自动改写，避免已发布链接失效。

const MAX_LEN = 80
/** 截断时至少保留这么长的前缀才回退到连字符处，否则直接硬截。 */
const MIN_DASH_INDEX = 20

type PinyinFn = (
  text: string,
  options?: { toneType?: 'none' | 'num' | 'symbol'; type?: 'array' | 'string'; nonZh?: 'consecutive' | 'spaced' | 'removed' }
) => string | string[]

let pinyinLoader: Promise<PinyinFn> | null = null

/** 拼音字典较大，按需加载，不进入前台首屏包。 */
function loadPinyin(): Promise<PinyinFn> {
  if (!pinyinLoader) {
    pinyinLoader = import('pinyin-pro').then((mod) => mod.pinyin as unknown as PinyinFn)
  }
  return pinyinLoader
}

/** 全角字母、数字与空格先转半角，避免「ＡＢＣ」「１２３」被整段丢掉。 */
function normalize(text: string): string {
  return text
    .replace(/[\uFF01-\uFF5E]/g, (ch) => String.fromCharCode(ch.charCodeAt(0) - 0xfee0))
    .replace(/\u3000/g, ' ')
}

/** 超过 80 字符时回退到最后一个连字符，不留半个拼音词。 */
function truncate(value: string): string {
  if (value.length <= MAX_LEN) return value
  const cut = value.slice(0, MAX_LEN)
  const dash = cut.lastIndexOf('-')
  return (dash >= MIN_DASH_INDEX ? cut.slice(0, dash) : cut).replace(/-+$/, '')
}

/** 标题 → slug：中文逐字转无声调拼音，英文与数字整段保留。 */
export async function slugify(title: string): Promise<string> {
  const text = normalize(title).trim()
  if (!text) return ''

  const pinyin = await loadPinyin()
  const tokens = (pinyin(text, { toneType: 'none', type: 'array', nonZh: 'consecutive' }) as string[])
    .map((item) => item.trim())
    .filter(Boolean)

  const cleaned = tokens
    .join('-')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/-{2,}/g, '-')
    .replace(/^-+|-+$/g, '')

  return truncate(cleaned)
}

/** 与已占用的标识去重：撞车时依次尝试 -2、-3…，后缀也算进 80 字符上限。 */
export function dedupeSlug(slug: string, taken: Iterable<string>): string {
  const used = taken instanceof Set ? taken : new Set(taken)
  if (!slug || !used.has(slug)) return slug

  for (let n = 2; n < 100; n += 1) {
    const suffix = `-${n}`
    const head = slug.slice(0, MAX_LEN - suffix.length).replace(/-+$/, '')
    const candidate = `${head}${suffix}`
    if (!used.has(candidate)) return candidate
  }
  return slug
}
