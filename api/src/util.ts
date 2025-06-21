export type Version = string | number;

/**
 * Compare two versions quickly.
 * @param current Is this version greater, equal to, or less than the other?
 * @param other The version to compare against the current version
 * @return 1 if current is greater than other, 0 if they are equal or equivalent, and -1 if current is less than other
 */
export default function versionCompare(
    current: Version,
    other: Version,
): number {
    const cp = String(current).split('.');
    const op = String(other).split('.');
    for (let depth = 0; depth < Math.min(cp.length, op.length); depth++) {
        const cn = Number(cp[depth]);
        const on = Number(op[depth]);
        if (cn > on) return 1;
        if (on > cn) return -1;
        if (!Number.isNaN(cn) && Number.isNaN(on)) return 1;
        if (Number.isNaN(cn) && !Number.isNaN(on)) return -1;
    }
    return 0;
}
