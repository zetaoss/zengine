export type Log = {
  level: string;
  args: unknown[];
};

export type Item = {
  arg: unknown;
  type: string;
  entries: Entry[] | null;
  name: string;
  needsExpand: boolean;
}

export type Entry = [string | number, unknown];

export function inspectArg(arg: unknown): Item {
  if (typeof arg === 'object') {
    if (arg === null) return newItem(arg, 'null');
    if (Array.isArray(arg)) return newItem(arg, 'Array', arg.map((item, index) => [index, item]));
    if (arg instanceof Map) return newItem(arg, 'Map', [...arg.entries()]);
    if (arg instanceof Set) return newItem(arg, 'Set', [...arg].map((item, index) => [index, item]));
    if (arg instanceof WeakMap) return newItem(arg, 'WeakMap');
    if (arg instanceof WeakSet) return newItem(arg, 'WeakSet');
    if (arg instanceof HTMLElement) return newItem(arg, 'HTMLElement');
    if (arg.constructor === undefined) {
      console.log('arg.constructor is undefined', arg);
      return newItem(arg, 'Object');
    }
    return newItem(arg, 'Object', Object.entries(arg), arg.constructor.name);
  }
  return newItem(arg, typeof arg);
}

function newItem(arg: unknown, type: string, entries: null | Entry[] = null, name: string = '', needsExpand: boolean = false): Item {
  return { arg, type, entries, name, needsExpand };
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
