export type Param = {
  arg: unknown;
  text: string;
  item: Item;
};

export type Log = {
  level: string;
  args: unknown[];
};

export type Item = {
  level: string;
  arg: unknown;
  type: string;
  key: Key;
  text: string;
  items: Item[];
  depth: number;
  circular: boolean;
}

type Entry = [string | number, unknown];
type Key = string | number | null;

export const getParam = (level: string, arg: unknown,): Param => {
  const item = createItem(level, arg);
  const text = stringify(arg);
  return { arg, text, item };
};

export const createItem = (level: string, arg: unknown, depth: number = 0, remaining: number = 5, key: Key = null, seen = new WeakSet()): Item => {
  let items: Item[] = [];
  const circular = false;
  const text = String(arg);
  if (arg === null || arg === undefined) {
    return { level, arg, type: text, depth, key, text, items, circular };
  }

  if (typeof arg !== 'object') {
    return { level, arg, type: typeof arg, depth, key, text, items, circular };
  }

  const [newCircular, type, entries] = inspectArg(arg, depth, key, seen);
  if (newCircular === false && remaining > 0) {
    items = entries.map(([k, v]) => createItem(level, v, depth + 1, remaining - 1, k, seen));
  }
  return { level, arg, type, depth, key, text, items, circular: newCircular };
};

export const getItemCircular = (level: string, arg: unknown, depth: number = 0, key: Key = null, seen = new WeakSet(), searchItems: boolean = true): Item => {
  const items: Item[] = [];
  const circular = false;
  const text = String(arg);
  if (arg === null || arg === undefined) {
    return { level, arg, type: String(arg), depth, key, text, items, circular };
  }

  if (typeof arg !== 'object') {
    return { level, arg, type: typeof arg, depth, key, text, items, circular };
  }

  const [newCircular, type, entries] = inspectArg(arg, depth, key, seen);
  if (searchItems) {
    const newItems = entries.map(([k, v]) => getItemCircular(level, v, depth + 1, k, seen, false));
    return { level, arg, type, depth, key, text, items: newItems, circular: newCircular };
  }
  return { level, arg, type, depth, key, text, items, circular: newCircular };
};


function inspectArg(arg: unknown, depth: number, key: Key, seen: WeakSet<object>): [boolean, string, Entry[]] {
  console.log(' '.repeat(depth * 4), key)
  // if (key == 'speechSynthesis') {
  //   return [false, 'speechSynthesis', []];
  // }
  if (typeof arg === 'object') {
    if (arg === null) return [false, 'null', []];
    if (seen.has(arg)) return [true, 'circular', []];
    seen.add(arg);
    if (Array.isArray(arg)) return [false, 'Array', arg.map((item, index) => [index, item])];
    if (arg instanceof Map) return [false, 'Map', [...arg.entries()]];
    if (arg instanceof Set) return [false, 'Set', [...arg].map((item, index) => [index, item])];
    if (arg instanceof WeakMap) return [false, 'WeakMap', []];
    if (arg instanceof WeakSet) return [false, 'WeakSet', []];
    if (arg instanceof HTMLElement) return [false, 'HTMLElement', []];
    if (arg.constructor === undefined) {
      console.log('arg.constructor is undefined', arg);
      return [false, 'Object', []];
    }
    const entries = Object.entries(arg)
    return [false, 'Object', entries];
    // if (arg.constructor.name === 'console') return [false, 'Object', Object.entries(arg).map(([fname]) => [fname, Function(`return function ${fname}(){CONSOLE}`)()])];
  }
  return [false, typeof arg, []];
}


export const stringify = (function () {
  const sortci = (a: string, b: string): number => {
    return a.toLowerCase() < b.toLowerCase() ? -1 : 1;
  };

  const htmlEntities = (str: string): string => {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  };

  return function stringify(
    o: unknown,
    visited: unknown[] = [],
    buffer: string = ""
  ): string {
    let type = "";
    const parts: string[] = [];

    if (o === null) return "null";
    if (typeof o === "undefined") return "undefined";
    if (typeof o === "number") return o.toString();
    if (typeof o === "boolean") return o ? "true" : "false";
    if (typeof o === "function") return o.toString().split("\n  ").join("\n" + buffer);
    if (typeof o === "string") return `"${htmlEntities(o.replace(/"/g, '\\"'))}"`;

    try {
      type = Object.prototype.toString.call(o);
    } catch {
      type = "[object Object]";
    }

    // Check for circular references
    if (visited.includes(o)) {
      return `[circular ${type.slice(1)}${o instanceof HTMLElement ? ` :\n${htmlEntities(o.outerHTML).split("\n").join("\n" + buffer)}` : ""}]`;
    }

    // Mark this object as visited
    visited.push(o);

    // Handle array case
    if (Array.isArray(o)) {
      for (const item of o) {
        parts.push(stringify(item, visited));
      }
      return `[${parts.join(", ")}]`;
    }

    // Handle generic object case
    const typeStr = `${type} `;
    const newBuffer = buffer + "  ";
    if (buffer.length / 2 < 2) {
      const names: string[] = [];
      try {
        for (const key in o) {
          names.push(key);
        }
      } catch { }

      names.sort(sortci);
      for (const name of names) {
        try {
          const value = o[name as keyof typeof o];
          parts.push(`${newBuffer}${name}: ${stringify(value, visited, newBuffer)}`);
        } catch { }
      }
    }

    return parts.length ? `${typeStr}{\n${parts.join(",\n")}\n${buffer}}` : `${typeStr}{ ... }`;
  };
})();
