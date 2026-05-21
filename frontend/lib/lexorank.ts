export class Lexorank {
  private readonly MIN_CHAR: number;
  private readonly MAX_CHAR: number;

  constructor() {
    this.MIN_CHAR = this.byte("0");
    this.MAX_CHAR = this.byte("z");
  }

  private byte(s: string): number {
    return s.charCodeAt(0);
  }

  private string(b: number): string {
    return String.fromCharCode(b);
  }

  private mid(prev: number, next: number): number {
    return Math.floor((prev + next) / 2);
  }

  private getChar(s: string, i: number, defaultChar: number): number {
    if (i >= s.length) {
      return defaultChar;
    }
    return this.byte(s.charAt(i));
  }

  // Returns [rank, ok]. ok=false signals space exhausted — trigger re-indexing.
  insert(p: string | null, n: string | null): [string, boolean] {
    if (p === "" || !p) {
      p = this.string(this.MIN_CHAR);
    }
    if (n === "" || !n) {
      n = this.string(this.MAX_CHAR);
    }

    let rank = "";
    let cursor = 0;

    while (true) {
      const prev_c = this.getChar(p, cursor, this.MIN_CHAR);
      const next_c = this.getChar(n, cursor, this.MAX_CHAR);

      if (prev_c === next_c) {
        rank += this.string(prev_c);
        cursor++;
        continue;
      }

      const mid_c = this.mid(prev_c, next_c);

      if (mid_c === prev_c || mid_c === next_c) {
        rank += this.string(prev_c);
        cursor++;
        continue;
      }

      rank += this.string(mid_c);
      break;
    }

    if (rank >= n) {
      return [p, false];
    }

    return [rank, true];
  }
}

export const compareLexorank = (a: string, b: string) =>
  a < b ? -1 : a > b ? 1 : 0;
