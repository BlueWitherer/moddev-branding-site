export class LogFormat {
    public static gray = 'color: gray';
    public static white = 'color: white';
    public static cyan = 'color: cyan';
    public static yellow = 'color: goldenrod';
    public static red = 'color: red';
    public static green = 'color: green';
    public static bold = 'font-weight: bold';
};

const timeStamp = (): string => {
    const now = new Date();
    const yyyy = now.getUTCFullYear();
    const mm = String(now.getUTCMonth() + 1).padStart(2, '0');
    const dd = String(now.getUTCDate()).padStart(2, '0');
    const h = String(now.getUTCHours()).padStart(2, '0');
    const m = String(now.getUTCMinutes()).padStart(2, '0');
    const s = String(now.getUTCSeconds()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd} ${h}:${m}:${s} UTC`;
};

const safeParseLog = (err: unknown): string => {
    if (err === null || err === undefined) return '[Null log]';
    try {
        return typeof err === 'object' ? JSON.stringify(err, null, 2) : String(err);
    } catch {
        return '[Unserializable log]';
    };
};

const formatArgs = (...args: any[]): string[] => args.map(safeParseLog);

const logMsg = (
    time: string,
    color: string,
    tag: string,
    ...args: any[]
): [string, ...string[]] => {
    const txt = args.join(' ');
    return [
        `%c${time}%c | %c${tag} | %c${txt}`,
        LogFormat.gray,
        color,
        `${LogFormat.bold};${color}`,
        color,
    ];
};

const originalConsole = {
    debug: console.debug,
    info: console.info,
    warn: console.warn,
    error: console.error,
    log: console.log,
};

/**
 * Pretty logging
 */
export default class log {
    public static debug = (...args: any[]): void => {
        originalConsole.debug(...logMsg(timeStamp(), LogFormat.gray, 'DEBUG', ...formatArgs(...args)));
    };

    public static info = (...args: any[]): void => {
        originalConsole.info(...logMsg(timeStamp(), LogFormat.cyan, 'INFO', ...formatArgs(...args)));
    };

    public static warn = (...args: any[]): void => {
        originalConsole.warn(...logMsg(timeStamp(), LogFormat.yellow, 'WARN', ...formatArgs(...args)));
    };

    public static error = (...args: any[]): void => {
        originalConsole.error(...logMsg(timeStamp(), LogFormat.red, 'ERROR', ...formatArgs(...args)));
    };

    public static done = (...args: any[]): void => {
        originalConsole.log(...logMsg(timeStamp(), LogFormat.green, 'DONE', ...formatArgs(...args)));
    };

    public static print = (...args: any[]): void => {
        originalConsole.log(...logMsg(timeStamp(), LogFormat.white, ' LOG ', ...formatArgs(...args)));
    };
};

// Override console methods
console.debug = (...args) => log.debug(...args);
console.info = (...args) => log.info(...args);
console.warn = (...args) => log.warn(...args);
console.error = (...args) => log.error(...args);
console.log = (...args) => log.print(...args);