import type * as Monaco from 'monaco-editor'

export const PROMQL_LANGUAGE_ID = 'promql'

// PromQL functions documentation
export const PROMQL_FUNCTIONS: Record<string, { signature: string; description: string }> = {
  // Aggregation operators
  sum: {
    signature: 'sum(v vector) vector',
    description: 'Calculate sum over dimensions'
  },
  avg: {
    signature: 'avg(v vector) vector',
    description: 'Calculate the average over dimensions'
  },
  min: {
    signature: 'min(v vector) vector',
    description: 'Select minimum over dimensions'
  },
  max: {
    signature: 'max(v vector) vector',
    description: 'Select maximum over dimensions'
  },
  count: {
    signature: 'count(v vector) vector',
    description: 'Count number of elements in the vector'
  },
  stddev: {
    signature: 'stddev(v vector) vector',
    description: 'Calculate population standard deviation over dimensions'
  },
  stdvar: {
    signature: 'stdvar(v vector) vector',
    description: 'Calculate population standard variance over dimensions'
  },
  topk: {
    signature: 'topk(k scalar, v vector) vector',
    description: 'Select largest k elements by sample value'
  },
  bottomk: {
    signature: 'bottomk(k scalar, v vector) vector',
    description: 'Select smallest k elements by sample value'
  },
  count_values: {
    signature: 'count_values(label string, v vector) vector',
    description: 'Count number of elements with the same value'
  },
  quantile: {
    signature: 'quantile(φ scalar, v vector) vector',
    description: 'Calculate φ-quantile (0 ≤ φ ≤ 1) over dimensions'
  },

  // Functions
  rate: {
    signature: 'rate(v range-vector) vector',
    description: 'Calculate the per-second average rate of increase of the time series in the range vector'
  },
  irate: {
    signature: 'irate(v range-vector) vector',
    description: 'Calculate the per-second instant rate of increase of the time series based on the last two data points'
  },
  increase: {
    signature: 'increase(v range-vector) vector',
    description: 'Calculate the increase in the time series in the range vector'
  },
  delta: {
    signature: 'delta(v range-vector) vector',
    description: 'Calculate the difference between the first and last value of each time series element'
  },
  idelta: {
    signature: 'idelta(v range-vector) vector',
    description: 'Calculate the difference between the last two samples'
  },
  deriv: {
    signature: 'deriv(v range-vector) vector',
    description: 'Calculate the per-second derivative using simple linear regression'
  },
  predict_linear: {
    signature: 'predict_linear(v range-vector, t scalar) vector',
    description: 'Predict the value of time series t seconds from now'
  },
  histogram_quantile: {
    signature: 'histogram_quantile(φ scalar, b vector) vector',
    description: 'Calculate the φ-quantile from a histogram'
  },

  // Math functions
  abs: {
    signature: 'abs(v vector) vector',
    description: 'Return absolute value'
  },
  ceil: {
    signature: 'ceil(v vector) vector',
    description: 'Round up to nearest integer'
  },
  floor: {
    signature: 'floor(v vector) vector',
    description: 'Round down to nearest integer'
  },
  round: {
    signature: 'round(v vector, to_nearest=1 scalar) vector',
    description: 'Round to nearest integer (or specified multiple)'
  },
  sqrt: {
    signature: 'sqrt(v vector) vector',
    description: 'Calculate square root'
  },
  exp: {
    signature: 'exp(v vector) vector',
    description: 'Calculate exponential function'
  },
  ln: {
    signature: 'ln(v vector) vector',
    description: 'Calculate natural logarithm'
  },
  log2: {
    signature: 'log2(v vector) vector',
    description: 'Calculate binary logarithm'
  },
  log10: {
    signature: 'log10(v vector) vector',
    description: 'Calculate decimal logarithm'
  },
  clamp: {
    signature: 'clamp(v vector, min scalar, max scalar) vector',
    description: 'Clamp samples to min/max values'
  },
  clamp_min: {
    signature: 'clamp_min(v vector, min scalar) vector',
    description: 'Clamp samples to minimum value'
  },
  clamp_max: {
    signature: 'clamp_max(v vector, max scalar) vector',
    description: 'Clamp samples to maximum value'
  },

  // Time functions
  time: {
    signature: 'time() scalar',
    description: 'Return the number of seconds since January 1, 1970 UTC'
  },
  timestamp: {
    signature: 'timestamp(v vector) vector',
    description: 'Return the timestamp of each sample'
  },
  day_of_month: {
    signature: 'day_of_month(v vector) vector',
    description: 'Return the day of the month for each sample timestamp (1-31)'
  },
  day_of_week: {
    signature: 'day_of_week(v vector) vector',
    description: 'Return the day of the week for each sample timestamp (0-6)'
  },
  day_of_year: {
    signature: 'day_of_year(v vector) vector',
    description: 'Return the day of the year for each sample timestamp (1-366)'
  },
  hour: {
    signature: 'hour(v vector) vector',
    description: 'Return the hour of the day for each sample timestamp (0-23)'
  },
  minute: {
    signature: 'minute(v vector) vector',
    description: 'Return the minute of the hour for each sample timestamp (0-59)'
  },
  month: {
    signature: 'month(v vector) vector',
    description: 'Return the month of the year for each sample timestamp (1-12)'
  },
  year: {
    signature: 'year(v vector) vector',
    description: 'Return the year for each sample timestamp'
  },

  // Label functions
  label_join: {
    signature: 'label_join(v vector, dst_label string, separator string, src_label_1 string, ...) vector',
    description: 'Join label values together'
  },
  label_replace: {
    signature: 'label_replace(v vector, dst_label string, replacement string, src_label string, regex string) vector',
    description: 'Replace label values with regex'
  },

  // Other functions
  absent: {
    signature: 'absent(v vector) vector',
    description: 'Return 1 if vector is empty, otherwise return nothing'
  },
  absent_over_time: {
    signature: 'absent_over_time(v range-vector) vector',
    description: 'Return 1 if range vector is empty, otherwise return nothing'
  },
  changes: {
    signature: 'changes(v range-vector) vector',
    description: 'Return number of times the value changed within the range'
  },
  resets: {
    signature: 'resets(v range-vector) vector',
    description: 'Return number of counter resets within the range'
  },
  sort: {
    signature: 'sort(v vector) vector',
    description: 'Sort by ascending sample value'
  },
  sort_desc: {
    signature: 'sort_desc(v vector) vector',
    description: 'Sort by descending sample value'
  },
  vector: {
    signature: 'vector(s scalar) vector',
    description: 'Return scalar as a vector with no labels'
  },
  scalar: {
    signature: 'scalar(v vector) scalar',
    description: 'Return single-element vector as scalar'
  }
}

// PromQL keywords
export const PROMQL_KEYWORDS = [
  'by', 'without', 'on', 'ignoring', 'group_left', 'group_right',
  'bool', 'offset', 'and', 'or', 'unless'
]

// PromQL operators
export const PROMQL_OPERATORS = [
  '+', '-', '*', '/', '%', '^',
  '==', '!=', '>', '<', '>=', '<=',
  '=~', '!~'
]

// Register PromQL language with Monaco
export function registerPromQLLanguage(monaco: typeof Monaco) {
  // Register the language
  monaco.languages.register({ id: PROMQL_LANGUAGE_ID })

  // Set language configuration
  monaco.languages.setLanguageConfiguration(PROMQL_LANGUAGE_ID, {
    comments: {
      lineComment: '#'
    },
    brackets: [
      ['{', '}'],
      ['[', ']'],
      ['(', ')']
    ],
    autoClosingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
      { open: "'", close: "'" }
    ],
    surroundingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
      { open: "'", close: "'" }
    ]
  })

  // Set tokenizer for syntax highlighting
  monaco.languages.setMonarchTokensProvider(PROMQL_LANGUAGE_ID, {
    keywords: PROMQL_KEYWORDS,
    functions: Object.keys(PROMQL_FUNCTIONS),

    tokenizer: {
      root: [
        // Comments
        [/#.*$/, 'comment'],

        // Strings
        [/"([^"\\]|\\.)*$/, 'string.invalid'], // non-terminated string
        [/'([^'\\]|\\.)*$/, 'string.invalid'], // non-terminated string
        [/"/, 'string', '@string_double'],
        [/'/, 'string', '@string_single'],

        // Numbers
        [/\d+(\.\d+)?([eE][+-]?\d+)?/, 'number'],

        // Duration literals
        [/\d+[smhdwy]/, 'number.duration'],

        // Operators
        [/[=!<>]=?|[+\-*/%^]|=~|!~/, 'operator'],

        // Brackets
        [/[{}()\[\]]/, '@brackets'],

        // Labels
        [/[a-zA-Z_][a-zA-Z0-9_]*(?=\s*[=!~])/, 'label'],

        // Functions and keywords
        [/[a-zA-Z_][a-zA-Z0-9_]*/, {
          cases: {
            '@keywords': 'keyword',
            '@functions': 'function',
            '@default': 'identifier'
          }
        }]
      ],

      string_double: [
        [/[^\\"]+/, 'string'],
        [/\\./, 'string.escape'],
        [/"/, 'string', '@pop']
      ],

      string_single: [
        [/[^\\']+/, 'string'],
        [/\\./, 'string.escape'],
        [/'/, 'string', '@pop']
      ]
    }
  })
}

// Define dark theme colors for PromQL
export function definePromQLTheme(monaco: typeof Monaco) {
  monaco.editor.defineTheme('promql-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [
      { token: 'comment', foreground: '6A9955', fontStyle: 'italic' },
      { token: 'string', foreground: 'CE9178' },
      { token: 'string.escape', foreground: 'D7BA7D' },
      { token: 'string.invalid', foreground: 'F44747' },
      { token: 'number', foreground: 'B5CEA8' },
      { token: 'number.duration', foreground: 'B5CEA8', fontStyle: 'bold' },
      { token: 'operator', foreground: 'D4D4D4' },
      { token: 'keyword', foreground: '569CD6', fontStyle: 'bold' },
      { token: 'function', foreground: 'DCDCAA' },
      { token: 'identifier', foreground: '9CDCFE' },
      { token: 'label', foreground: '4EC9B0' }
    ],
    colors: {
      'editor.background': '#1e1e2e',
      'editor.foreground': '#cdd6f4',
      'editor.lineHighlightBackground': '#313244',
      'editorCursor.foreground': '#f5e0dc',
      'editor.selectionBackground': '#45475a',
      'editorLineNumber.foreground': '#6c7086',
      'editorLineNumber.activeForeground': '#cdd6f4'
    }
  })
}
