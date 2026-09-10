export { cn } from "cn";

/** ISO date-time string -> "YYYY-MM-DD" for an <input type="date">. */
export function toDateInputValue(iso?: string | null): string {
  if (!iso) return "";
  return iso.slice(0, 10);
}

/**
 * "YYYY-MM-DD" from an <input type="date"> -> ISO date-time string, or undefined if empty.
 * Anchored at UTC midnight (not local time) so the calendar date round-trips correctly
 * through toDateInputValue() regardless of the viewer's timezone.
 */
export function fromDateInputValue(value: string): string | undefined {
  if (!value) return undefined;
  return `${value}T00:00:00.000Z`;
}

/** Today's date as "YYYY-MM-DD" in the viewer's local timezone. */
export function todayDateString(): string {
  return new Date().toLocaleDateString("sv-SE");
}

/** Add days to a "YYYY-MM-DD" string, returned as "YYYY-MM-DD" (local calendar arithmetic). */
export function addDaysToDateString(date: string, days: number): string {
  const d = new Date(`${date}T00:00:00`);
  d.setDate(d.getDate() + days);
  return d.toLocaleDateString("sv-SE");
}
