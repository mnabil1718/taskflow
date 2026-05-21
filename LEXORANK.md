## Lexorank Algorithm Implementation in TypeScript

```ts
/**
 * Implementation of Atlassion Lexicography Ranking System
 * see: https://support.atlassian.com/jira/kb/understanding-and-managing-lexorank-in-jira-server
 *
 * ref: https://medium.com/turkcell/lexorank-managing-sorted-tables-with-ease-f404f7eb00a9
 */
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

  /**
   *
   * Calculate middle code given previous and next character codes
   */
  private mid(prev: number, next: number): number {
    return Math.floor((prev + next) / 2);
  }

  /**
   * Get character code for a given string
   * character at a given index
   */
  private getChar(s: string, i: number, defaultChar: number): number {
    if (i >= s.length) {
      return defaultChar;
    }

    return this.byte(s.charAt(i));
  }

  /**
   *
   * calculate new rank given p (previous) string value
   * and n (next) string value. Previous & Next can refer to
   * neighbouring item around the inserted item.
   *
   * if p is null, meaning it is the first item in the list.
   * if n is null. it is the last item in the list.
   *
   * returns the rank produced and ok boolean value. OK value
   * can be used to trigger re-indexing procedure.
   *
   */
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

      // prev_c and next_c is probably adjacent ("touching"). e.g. a and b
      if (mid_c === prev_c || mid_c === next_c) {
        rank += this.string(prev_c);
        cursor++;
        continue;
      }

      // if not append mid_c
      rank += this.string(mid_c);
      break;
    }

    // space exhausted, ok=false
    // tells caller to trigger re-indexing
    if (rank >= n) {
      return [p, false];
    }

    return [rank, true];
  }
}

// utility function for sorting
export const compareLexorank = (a: string, b: string) =>
  a < b ? -1 : a > b ? 1 : 0;
```
