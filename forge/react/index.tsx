import { useId, useLayoutEffect, useRef } from 'react';
import type { ComponentPropsWithRef, ReactNode } from 'react';

type Size = 'xsmall' | 'small' | 'medium' | 'large' | 'xlarge';
type Shape = 'round' | 'square';
type NativeButton = ComponentPropsWithRef<'button'>;
type ButtonAppearance = { size?: Size; shape?: Shape; pressed?: boolean };

export type ButtonProps = NativeButton & ButtonAppearance & {
  variant?: 'filled' | 'tonal' | 'outlined' | 'text' | 'elevated';
  color?: 'primary' | 'secondary' | 'tertiary' | 'error';
};

export function Button({
  variant = 'filled', size = 'small', shape = 'round', type = 'button',
  color, pressed, className = '', ...props
}: ButtonProps) {
  return <button {...props} type={type} className={`md-button md-${variant}-button ${className}`.trim()}
    data-size={size} data-shape={shape} data-color={color} aria-pressed={pressed ?? props['aria-pressed']} />;
}

export type ButtonLinkProps = Omit<ComponentPropsWithRef<'a'>, 'color'> &
  Pick<ButtonProps, 'variant' | 'color' | 'size' | 'shape'> & { href: string };

export function ButtonLink({
  variant = 'filled', size = 'small', shape = 'round', color, className = '', ...props
}: ButtonLinkProps) {
  return <a {...props} className={`md-button md-${variant}-button ${className}`.trim()}
    data-size={size} data-shape={shape} data-color={color} />;
}

export type IconButtonProps = Omit<NativeButton, 'aria-label'> & ButtonAppearance & {
  label: string;
  variant?: 'standard' | 'filled' | 'tonal' | 'outlined';
  width?: 'narrow' | 'default' | 'wide';
  children: ReactNode;
};

export function IconButton({
  variant = 'standard', size = 'small', shape = 'round', width = 'default',
  type = 'button', label, pressed, className = '', ...props
}: IconButtonProps) {
  return <button {...props} type={type} className={`md-icon-button md-${variant}-icon-button ${className}`.trim()}
    data-size={size} data-shape={shape} data-width={width} aria-label={label}
    aria-pressed={pressed ?? props['aria-pressed']} />;
}

export type ListProps = ComponentPropsWithRef<'ul'>;

export function List({ className = '', ...props }: ListProps) {
  return <ul data-variant="standard" data-shape="expressive" data-size="one-line"
    data-width="icon" data-divider="none" {...props} className={`md-list ${className}`.trim()} />;
}

export function SearchBar({ label, leading, action, placeholder = label, className = '', ...props }: ComponentPropsWithRef<'input'> & {
  label: string; leading: ReactNode; action?: ReactNode;
}) {
  return <div className={`md-search-bar ${className}`.trim()} data-control="contained" data-leading="icon">
    <span className="md-search-leading-slot"><span className="md-search-leading" aria-hidden="true">{leading}</span></span>
    <input {...props} type="search" className="md-search-input" aria-label={label} placeholder={placeholder} />
    {action != null && <div className="md-search-actions"><span className="md-search-action-target">{action}</span></div>}
  </div>;
}

export function SplitButton({ label, size = 'small', variant = 'tonal', primary, menu, className = '' }: {
  label: string; size?: ButtonProps['size']; variant?: ButtonProps['variant']; primary: ReactNode; menu: ReactNode; className?: string;
}) {
  return <div role="group" aria-label={label} className={`md-split-button ${className}`.trim()}
    data-variant={variant} data-size={size} data-shape="icon-label">{primary}{menu}</div>;
}

export function Fab({ label, size = 'default', variant = 'primary-container', className = '', ...props }: NativeButton & {
  label: string; size?: 'small' | 'default' | 'medium' | 'large'; variant?: 'primary-container' | 'secondary-container' | 'tertiary-container';
}) {
  return <button {...props} type="button" aria-label={label} className={`md-button md-fab ${className}`.trim()}
    data-size={size} data-variant={variant} data-width="default" />;
}

export function NavigationRail({ label, children, action, footer, width = 'narrow', className = '' }: {
  label: string; children: ReactNode; action?: ReactNode; footer?: ReactNode; width?: 'default' | 'narrow'; className?: string;
}) {
  return <nav aria-label={label} className={`md-navigation md-navigation-rail ${className}`.trim()}
    data-variant="collapsed" data-width={width} data-orientation="vertical" data-alignment="top"
    data-layout="standard" data-indicator="fill" data-size="default">
    {action != null && <header>{action}</header>}
    <div className="md-nav-items">{children}</div>
    {footer != null && <footer>{footer}</footer>}
  </nav>;
}

export function NavigationRailItem({ label, icon, ...props }: NativeButton & { label: string; icon: ReactNode }) {
  return <button {...props} type="button" className="md-nav-item" data-navigation-item="">
    <span className="md-nav-indicator">
      <span className="md-nav-icon" aria-hidden="true"><span style={{ display: 'block' }}>{icon}</span></span>
      <span className="md-nav-label" data-navigation-label={label}><span>{label}</span></span>
    </span>
  </button>;
}

export function SideSheet({ title, headerAction, children, variant = 'docked', className = '', ...props }: ComponentPropsWithRef<'aside'> & {
  title: string; headerAction?: ReactNode; variant?: 'docked' | 'detached';
}) {
  return <aside {...props} aria-label={title} className={`md-sheet md-side-sheet ${className}`.trim()}
    data-control="standard" data-variant={variant} data-orientation="end">
    <header className="md-sheet-header"><h2 className="md-sheet-title">{title}</h2>{headerAction}</header>
    <div className="md-sheet-body">{children}</div>
  </aside>;
}

export type ListItemButtonProps = NativeButton & { leading?: ReactNode; trailing?: ReactNode; actions?: ReactNode };

export function ListItemButton({ leading, trailing, actions, children, className = '', type = 'button', ...props }: ListItemButtonProps) {
  const content = <>
    {leading != null && <span className="md-list-leading md-list-icon" aria-hidden="true">{leading}</span>}
    <span className="md-list-body"><span className="md-list-label">{children}</span></span>
    {trailing != null && <span className="md-list-tail">{trailing}</span>}
  </>;
  const button = <button {...props} type={type} className={`md-list-item ${className}`.trim()} data-list-item="">
    {actions == null ? content : <span className="md-list-main">{content}</span>}
  </button>;
  return actions == null ? button : <div className="md-list-row">
    {button}
    <div className="md-list-actions">{actions}</div>
  </div>;
}

export type CheckboxProps = Omit<ComponentPropsWithRef<'input'>, 'type' | 'children'>;

export function Checkbox(props: CheckboxProps) {
  return <label className="md-checkbox">
    <input {...props} type="checkbox" />
    <span className="md-checkbox-box" aria-hidden="true">
      <svg className="md-checkbox-check" viewBox="0 0 18 18" fill="none">
        <path d="m4 9 3 3 7-7" stroke="currentColor" strokeWidth="2" />
      </svg>
      <svg className="md-checkbox-dash" viewBox="0 0 18 18" fill="none">
        <path d="M4 9h10" stroke="currentColor" strokeWidth="2" />
      </svg>
    </span>
  </label>;
}

export type TextFieldProps = ComponentPropsWithRef<'input'> & {
  label: string;
  variant?: 'outlined' | 'filled';
  invalid?: boolean;
  supporting?: ReactNode;
  leading?: ReactNode;
  trailing?: ReactNode;
  containerClassName?: string;
};

export function TextField({
  label, variant = 'outlined', invalid, supporting, leading, trailing, containerClassName = '',
  id, placeholder = ' ', ...props
}: TextFieldProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const invalidState = invalid ?? props['aria-invalid'] ?? false;
  const description = [props['aria-describedby'], supporting != null ? `${inputId}-help` : null].filter(Boolean).join(' ');
  return <div className={`md-${variant}-text-field ${containerClassName}`.trim()} data-invalid={invalidState}>
    {variant === 'outlined' && <>
      <fieldset className="md-field-outline" aria-hidden="true"><legend>{label}</legend></fieldset>
      <fieldset className="md-field-outline md-field-focus-outline" aria-hidden="true"><legend>{label}</legend></fieldset>
    </>}
    <label htmlFor={inputId} data-slot="label">{label}</label>
    {leading != null && <span data-slot="leading-icon" aria-hidden="true">{leading}</span>}
    <input {...props} id={inputId} placeholder={placeholder} aria-invalid={invalidState} aria-describedby={description || undefined} />
    {trailing != null && <span data-slot="trailing-icon">{trailing}</span>}
    {supporting != null && <small data-slot="supporting-text" id={`${inputId}-help`}>{supporting}</small>}
  </div>;
}

type DialogController = {
  open(panel: HTMLDialogElement, opener?: HTMLElement): Promise<void>;
  close(panel: HTMLDialogElement, result: string, reason: string): Promise<void>;
  dispose(panel: HTMLDialogElement): void;
};

const dialogController = (globalThis as typeof globalThis & { MaterialDialogs?: DialogController }).MaterialDialogs;

export type DialogProps = Omit<ComponentPropsWithRef<'dialog'>, 'ref' | 'onClose' | 'onCancel' | 'title'> & {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAfterClose?: () => void;
  opener?: HTMLElement | null;
  title: ReactNode;
  headerAction?: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
};

export function Dialog({
  open, onOpenChange, onAfterClose, opener, title, headerAction, description, actions, children,
  className = '', ...props
}: DialogProps) {
  const id = useId();
  const ref = useRef<HTMLDialogElement>(null);

  useLayoutEffect(() => {
    const panel = ref.current!;
    return () => dialogController?.dispose(panel);
  }, []);

  useLayoutEffect(() => {
    const panel = ref.current!;
    // Forge detects dismissal. The React owner accepts it by changing open.
    function requestClose(event: Event) {
      const { reason } = (event as CustomEvent<{ reason: string }>).detail;
      if (reason === 'controlled') return;
      event.preventDefault();
      onOpenChange(false);
    }
    function afterClose() {
      if (open) onOpenChange(false);
      onAfterClose?.();
    }
    panel.addEventListener('md-dialog-beforeclose', requestClose);
    panel.addEventListener('md-dialog-closed', afterClose);
    return () => {
      panel.removeEventListener('md-dialog-beforeclose', requestClose);
      panel.removeEventListener('md-dialog-closed', afterClose);
    };
  }, [open, onOpenChange, onAfterClose]);

  useLayoutEffect(() => {
    const panel = ref.current!;
    if (open) void dialogController?.open(panel, opener ?? undefined);
    else void dialogController?.close(panel, '', 'controlled');
  }, [open, opener]);

  return <dialog {...props} ref={ref} className={`md-dialog ${className}`.trim()} data-dialog-surface="" data-variant="basic"
    aria-labelledby={`${id}-heading`} aria-describedby={description != null ? `${id}-support` : undefined}>
    <header className="md-basic-header"><h2 id={`${id}-heading`}>{title}</h2>{headerAction}</header>
    <div className="md-dialog-content" data-dialog-content="">
      {description != null && <p data-slot="supporting-text" id={`${id}-support`}>{description}</p>}
      {children}
    </div>
    {actions != null && <footer className="md-basic-actions">{actions}</footer>}
  </dialog>;
}
