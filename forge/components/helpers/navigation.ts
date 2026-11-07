// Stateless keyboard navigation shared by independently delivered controls.
export function navigationIndex(index: number, length: number, key: string, horizontal: boolean, rtl = false) {
  if (key === "Home") return 0;
  if (key === "End") return length - 1;
  const forward = horizontal
    ? rtl ? "ArrowLeft" : "ArrowRight"
    : "ArrowDown";
  const backward = horizontal
    ? rtl ? "ArrowRight" : "ArrowLeft"
    : "ArrowUp";
  if (key !== forward && key !== backward) return null;
  return (index + (key === forward ? 1 : -1) + length) % length;
}
