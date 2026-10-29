declare module 'swagger-ui-dist/swagger-ui-es-bundle.js' {
  export interface SwaggerUIInstance {
    destroy?: () => void
  }

  interface SwaggerUIOptions {
    dom_id: string
    url: string
    deepLinking?: boolean
    displayRequestDuration?: boolean
    docExpansion?: 'none' | 'list' | 'full'
    tagsSorter?: 'alpha'
    tryItOutEnabled?: boolean
    validatorUrl?: string | null
    withCredentials?: boolean
  }

  export default function SwaggerUIBundle(
    options: SwaggerUIOptions,
  ): SwaggerUIInstance
}
