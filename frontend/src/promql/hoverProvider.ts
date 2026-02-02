import type * as Monaco from 'monaco-editor'
import { PROMQL_FUNCTIONS, PROMQL_KEYWORDS, PROMQL_LANGUAGE_ID } from './language'

// Create hover provider for function documentation
export function createHoverProvider(monaco: typeof Monaco): Monaco.languages.HoverProvider {
  return {
    provideHover(
      model: Monaco.editor.ITextModel,
      position: Monaco.Position
    ): Monaco.languages.ProviderResult<Monaco.languages.Hover> {
      const word = model.getWordAtPosition(position)
      if (!word) return null

      const wordText = word.word.toLowerCase()

      // Check if it's a function
      const funcInfo = PROMQL_FUNCTIONS[wordText]
      if (funcInfo) {
        return {
          range: {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endColumn: word.endColumn
          },
          contents: [
            { value: `**${wordText}**` },
            { value: '```\n' + funcInfo.signature + '\n```' },
            { value: funcInfo.description }
          ]
        }
      }

      // Check if it's a keyword
      if (PROMQL_KEYWORDS.includes(wordText)) {
        const keywordDescriptions: Record<string, string> = {
          by: 'Preserve the listed labels in the result',
          without: 'Remove the listed labels from the result',
          on: 'Match labels for binary operators',
          ignoring: 'Ignore listed labels when matching',
          group_left: 'Many-to-one matching (keep left side labels)',
          group_right: 'One-to-many matching (keep right side labels)',
          bool: 'Return 0/1 instead of filtering for comparison operators',
          offset: 'Time offset for lookback',
          and: 'Intersection of two vectors',
          or: 'Union of two vectors',
          unless: 'Complement of two vectors'
        }

        const description = keywordDescriptions[wordText]
        if (description) {
          return {
            range: {
              startLineNumber: position.lineNumber,
              endLineNumber: position.lineNumber,
              startColumn: word.startColumn,
              endColumn: word.endColumn
            },
            contents: [
              { value: `**${wordText}** (keyword)` },
              { value: description }
            ]
          }
        }
      }

      return null
    }
  }
}

// Register the hover provider
export function registerHoverProvider(monaco: typeof Monaco) {
  monaco.languages.registerHoverProvider(
    PROMQL_LANGUAGE_ID,
    createHoverProvider(monaco)
  )
}
