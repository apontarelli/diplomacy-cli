// HTMX TypeScript definitions for modern usage
declare global {
  interface Window {
    htmx: {
      // Core HTMX functions
      ajax: (verb: string, path: string, context?: any) => void;
      find: (selector: string) => Element | null;
      findAll: (selector: string) => NodeListOf<Element>;
      trigger: (element: Element | string, event: string, detail?: any) => void;
      on: (event: string, handler: (evt: CustomEvent) => void) => void;
      off: (event: string, handler: (evt: CustomEvent) => void) => void;
      
      // HTMX 2.0 specific features
      swap: (target: Element, content: string, swapSpec?: string) => void;
      settle: (element: Element) => void;
      
      // Configuration
      config: {
        historyEnabled: boolean;
        historyCacheSize: number;
        refreshOnHistoryMiss: boolean;
        defaultSwapStyle: string;
        defaultSwapDelay: number;
        defaultSettleDelay: number;
        includeIndicatorStyles: boolean;
        indicatorClass: string;
        requestClass: string;
        addedClass: string;
        settlingClass: string;
        swappingClass: string;
        allowEval: boolean;
        allowScriptTags: boolean;
        inlineScriptNonce: string;
        attributesToSettle: string[];
        wsReconnectDelay: string;
        wsBinaryType: string;
        disableSelector: string;
        withCredentials: boolean;
        timeout: number;
        scrollBehavior: string;
        defaultFocusScroll: boolean;
        getCacheBusterParam: boolean;
        globalViewTransitions: boolean;
        methodsThatUseUrlParams: string[];
        selfRequestsOnly: boolean;
        ignoreTitle: boolean;
        scrollIntoViewOnBoost: boolean;
        triggerSpecsCache: any;
      };
      
      // Logging and debugging
      logAll: () => void;
      logNone: () => void;
      logger: (element: Element, event: string, data: any) => void;
      
      // Extensions
      defineExtension: (name: string, extension: any) => void;
      removeExtension: (name: string) => void;
      
      // Processing
      process: (element: Element) => void;
      
      // Values and includes
      values: (element: Element, type?: string) => any;
      
      // WebSocket extension (when loaded)
      ws?: {
        send: (element: Element, message: any) => void;
        createSocket: (element: Element) => WebSocket | null;
      };
    };
  }

  // Custom events that HTMX fires
  interface HTMLElementEventMap {
    'htmx:configRequest': CustomEvent<{
      parameters: Record<string, any>;
      unfilteredParameters: Record<string, any>;
      headers: Record<string, string>;
      xhr: XMLHttpRequest;
    }>;
    'htmx:confirm': CustomEvent<{
      target: Element;
      triggeringEvent: Event;
      question: string;
      issueRequest: (skipConfirmation?: boolean) => void;
    }>;
    'htmx:historyCacheError': CustomEvent<{
      xhr: XMLHttpRequest;
      pathInfo: any;
    }>;
    'htmx:historyCacheMiss': CustomEvent<{
      xhr: XMLHttpRequest;
      pathInfo: any;
    }>;
    'htmx:historyCacheMissError': CustomEvent<{
      xhr: XMLHttpRequest;
      pathInfo: any;
    }>;
    'htmx:historyCacheMissLoad': CustomEvent<{
      xhr: XMLHttpRequest;
      pathInfo: any;
    }>;
    'htmx:historyRestore': CustomEvent<{
      path: string;
      cacheMiss: boolean;
    }>;
    'htmx:beforeCleanupElement': CustomEvent<{
      target: Element;
    }>;
    'htmx:load': CustomEvent<{
      target: Element;
    }>;
    'htmx:noSSESourceError': CustomEvent<{
      target: Element;
    }>;
    'htmx:onLoadError': CustomEvent<{
      target: Element;
      exception: Error;
    }>;
    'htmx:oobAfterSwap': CustomEvent<{
      target: Element;
    }>;
    'htmx:oobBeforeSwap': CustomEvent<{
      target: Element;
    }>;
    'htmx:oobErrorNoTarget': CustomEvent<{
      target: Element;
    }>;
    'htmx:prompt': CustomEvent<{
      target: Element;
      triggeringEvent: Event;
      message: string;
      prompt: string;
    }>;
    'htmx:pushedIntoHistory': CustomEvent<{
      path: string;
    }>;
    'htmx:responseError': CustomEvent<{
      xhr: XMLHttpRequest;
      target: Element;
    }>;
    'htmx:sendError': CustomEvent<{
      xhr: XMLHttpRequest;
      target: Element;
    }>;
    'htmx:sseError': CustomEvent<{
      target: Element;
      error: Event;
      source: EventSource;
    }>;
    'htmx:sseOpen': CustomEvent<{
      target: Element;
      source: EventSource;
    }>;
    'htmx:swapError': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    'htmx:targetError': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    'htmx:timeout': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    'htmx:validation:validate': CustomEvent<{
      target: Element;
      message: string;
      validity: ValidityState;
    }>;
    'htmx:validation:failed': CustomEvent<{
      target: Element;
      message: string;
      validity: ValidityState;
    }>;
    'htmx:validation:halted': CustomEvent<{
      target: Element;
      errors: any[];
    }>;
    'htmx:xhr:abort': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    'htmx:xhr:loadend': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    'htmx:xhr:loadstart': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    'htmx:xhr:progress': CustomEvent<{
      target: Element;
      xhr: XMLHttpRequest;
    }>;
    
    // WebSocket events (when ws extension is loaded)
    'htmx:wsConnecting': CustomEvent<{
      target: Element;
    }>;
    'htmx:wsOpen': CustomEvent<{
      target: Element;
      socket: WebSocket;
    }>;
    'htmx:wsClose': CustomEvent<{
      target: Element;
      socket: WebSocket;
    }>;
    'htmx:wsError': CustomEvent<{
      target: Element;
      error: Event;
      socket: WebSocket;
    }>;
    'htmx:wsBeforeMessage': CustomEvent<{
      target: Element;
      message: string;
      socket: WebSocket;
    }>;
    'htmx:wsAfterMessage': CustomEvent<{
      target: Element;
      message: string;
      socket: WebSocket;
    }>;
  }
}

export {};