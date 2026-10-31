import { createContext, useContext, useEffect, useState, useMemo, type ReactNode } from 'react'
import { getCookie, setCookie } from '@/lib/cookies'
import { translations, type Language, type TranslationSchema } from '@/locales'

const LANG_STORAGE_KEY = 'polyglot_lang'
const LANG_COOKIE_MAX_AGE = 60 * 60 * 24 * 365 // 1 year
const DEFAULT_LANGUAGE: Language = 'id'

type NestedKeyOf<ObjectType extends object> = {
  [Key in keyof ObjectType & (string | number)]: ObjectType[Key] extends object
    ? `${Key}.${NestedKeyOf<ObjectType[Key]>}`
    : `${Key}`
}[keyof ObjectType & (string | number)]

export type TranslationKey = NestedKeyOf<TranslationSchema>

interface I18nContextState {
  language: Language
  setLanguage: (lang: Language) => void
  t: (key: TranslationKey | string, params?: Record<string, string | number>) => string
}

const I18nContext = createContext<I18nContextState | null>(null)

interface I18nProviderProps {
  children: ReactNode
  defaultLanguage?: Language
}

export function I18nProvider({ children, defaultLanguage }: I18nProviderProps) {
  const [language, _setLanguage] = useState<Language>(() => {
    const saved = (getCookie(LANG_STORAGE_KEY) || localStorage.getItem(LANG_STORAGE_KEY)) as Language
    if (saved === 'en' || saved === 'id') return saved
    return defaultLanguage || DEFAULT_LANGUAGE
  })

  useEffect(() => {
    document.documentElement.lang = language
  }, [language])

  const setLanguage = (lang: Language) => {
    setCookie(LANG_STORAGE_KEY, lang, LANG_COOKIE_MAX_AGE)
    try {
      localStorage.setItem(LANG_STORAGE_KEY, lang)
    } catch {
      // Ignored for restricted storage
    }
    _setLanguage(lang)
  }

  const currentDictionary = useMemo(() => translations[language] || translations.en, [language])
  const fallbackDictionary = translations.en

  const t = useMemo(() => {
    return (key: TranslationKey | string, params?: Record<string, string | number>): string => {
      const parts = key.split('.')
      let value: any = currentDictionary
      for (const part of parts) {
        if (value && typeof value === 'object' && part in value) {
          value = value[part]
        } else {
          value = undefined
          break
        }
      }

      // Fallback to English if missing in current language
      if (value === undefined) {
        let fallbackVal: any = fallbackDictionary
        for (const part of parts) {
          if (fallbackVal && typeof fallbackVal === 'object' && part in fallbackVal) {
            fallbackVal = fallbackVal[part]
          } else {
            fallbackVal = undefined
            break
          }
        }
        value = fallbackVal
      }

      if (typeof value !== 'string') {
        return key
      }

      if (params) {
        return Object.entries(params).reduce((str, [paramKey, paramVal]) => {
          return str.split(`{${paramKey}}`).join(String(paramVal))
        }, value)
      }

      return value
    }
  }, [currentDictionary, fallbackDictionary])

  const contextValue = useMemo(() => ({
    language,
    setLanguage,
    t,
  }), [language, t])

  return (
    <I18nContext.Provider value={contextValue}>
      {children}
    </I18nContext.Provider>
  )
}

export const useI18n = () => {
  const context = useContext(I18nContext)
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider')
  }
  return context
}
