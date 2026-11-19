import { Dialog, Popover } from 'react-aria-components';
import type { ReactNode, RefObject } from 'react';

export function AnchoredDialog({ anchor, label, onDismiss, children, className = '' }: {
  anchor: RefObject<HTMLElement | null>;
  label: string;
  onDismiss: () => void;
  children: ReactNode;
  className?: string;
}) {
  return <Popover triggerRef={anchor} isOpen onOpenChange={open => { if (!open) onDismiss(); }}
    placement="bottom start" offset={6} containerPadding={12}
    className={`editor-popover ${className}`.trim()} data-component-assets>
    <Dialog aria-label={label}>{children}</Dialog>
  </Popover>;
}
