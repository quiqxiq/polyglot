import { en } from './en'
import { id } from './id'
import { type Language, type TranslationSchema } from './types'

export const translations: Record<Language, TranslationSchema> = {
  en,
  id,
}

export * from './types'
export { en, id }
