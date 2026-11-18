import { Children, isValidElement, useId, useImperativeHandle, useRef } from 'react';
import type { ComponentProps, ComponentPropsWithRef, ReactNode } from 'react';

export function SelectOption(props: ComponentProps<'option'>) {
  return <option {...props} />;
}

type OptionProps = ComponentProps<'option'>;
type OptionGroup = { label?: string; options: OptionProps[]; disabled?: boolean };
type FieldProps = Pick<ComponentProps<'select'>, 'name' | 'required' | 'value' | 'onChange' | 'disabled' | 'children'> & {
  id?: string;
  className?: string;
};
export type SelectProps = FieldProps & Omit<ComponentPropsWithRef<'input'>,
  keyof FieldProps | 'type' | 'defaultValue' | 'readOnly'>;

function optionGroups(children: ReactNode): OptionGroup[] {
  const groups: OptionGroup[] = [];
  Children.forEach(children, child => {
    if (!isValidElement<OptionProps>(child)) return;
    if (child.type === 'optgroup') {
      const { label, disabled, children: options } = child.props;
      groups.push({ label, disabled, options: optionGroups(options).flatMap(group => group.options) });
    } else {
      const previous = groups.at(-1);
      if (previous && previous.label === undefined) previous.options.push(child.props);
      else groups.push({ options: [child.props] });
    }
  });
  return groups;
}

function optionLabel(option: OptionProps) {
  return option.label ?? Children.toArray(option.children).join('');
}

function optionValue(option: OptionProps) {
  return String(option.value ?? optionLabel(option));
}

function Chevron() {
  return <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" aria-hidden="true">
    <path d="m7.4 8.6 4.6 4.6 4.6-4.6L18 10l-6 6-6-6z" />
  </svg>;
}

function Check() {
  return <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" aria-hidden="true">
    <path d="m9 16.17-4.17-4.17-1.42 1.41L9 19 21 7l-1.41-1.41z" />
  </svg>;
}

function FieldSelect({ chip, selectedLabel, inputProps, className = '', id, name, required, value, onChange, disabled, children }: FieldProps & {
  chip?: string;
  selectedLabel?: string;
  inputProps?: Omit<SelectProps, keyof FieldProps>;
}) {
  const uid = useId();
  const controlId = id ?? uid;
  const listId = uid + '-options';
  const trigger = useRef<HTMLInputElement>(null);
  const { ref, ...attributes } = inputProps ?? {};
  useImperativeHandle(ref, () => trigger.current!, []);
  const groups = optionGroups(children);
  const current = groups.flatMap(group => group.options).find(option => optionValue(option) === String(value ?? ''));
  const display = current ? optionLabel(current) : String(value ?? '');
  const base = { id: controlId, role: 'combobox', 'aria-expanded': false, 'aria-haspopup': 'listbox' as const,
    'aria-controls': listId, 'aria-autocomplete': 'none' as const, disabled, 'data-field-combo': '', 'data-select-value': value };

  return <div className={chip ? 'md-chip-select ' + className : 'md-outlined-select md-react-field-select ' + className}
    data-field-example data-react-field data-populated="true">
    {chip ? <button {...base} type="button" className="md-filter-chip" aria-label={chip} data-field-select="">
      <span data-field-display>{selectedLabel ?? display}</span>
      <span className="md-chip-icon" data-slot="trailing-icon"><Chevron /></span>
    </button> : <>
      <input {...attributes} {...base} ref={trigger} type="text" readOnly value={display}
        aria-required={required || undefined} />
      <button type="button" className="md-field-toggle" data-field-toggle="" tabIndex={-1}
        aria-label="Show options" aria-controls={listId} disabled={disabled}>
        <span data-slot="trailing-icon"><Chevron /></span>
      </button>
    </>}
    <select hidden tabIndex={-1} name={name} required={required} disabled={disabled} value={value} onChange={onChange}
      data-field-value="" onInvalid={event => {
        event.preventDefault();
        trigger.current?.focus();
      }}>{children}</select>
    <div id={listId} className="md-field-menu md-menu" data-variant="standard" data-orientation="vertical"
      role="listbox" aria-labelledby={attributes['aria-labelledby']}
      aria-label={chip ?? attributes['aria-label']} popover="auto">
      <section>
        {groups.map(({ label, options, disabled: groupDisabled }, groupIndex) =>
          <div role={label ? 'group' : 'presentation'} aria-label={label} key={groupIndex}>
            {label && <h3 aria-hidden="true">{label}</h3>}
            <ul role="presentation">
              {options.map((option, index) => {
                const key = optionValue(option);
                return <li role="presentation" key={key}>
                  <div className="md-menu-item" role="option" id={listId + '-' + groupIndex + '-' + index}
                    aria-selected={key === String(value ?? '')} aria-disabled={!!(groupDisabled || option.disabled)}
                    data-field-value={key} data-field-option={optionLabel(option)}>
                    <span data-option-label>{option.label ?? option.children}</span>
                    <span data-slot="trailing-icon" aria-hidden="true"><Check /></span>
                  </div>
                </li>;
              })}
            </ul>
          </div>)}
      </section>
      <p className="md-field-empty" role="option" aria-disabled="true" hidden>No matches</p>
    </div>
  </div>;
}

export function Select({ name, required, value, onChange, disabled, children, className, id, ...inputProps }: SelectProps) {
  return <FieldSelect {...{ name, required, value, onChange, disabled, children, className, id, inputProps }} />;
}

export function ChipSelect({ label, ...props }: FieldProps & { label: string; selectedLabel: string }) {
  return <FieldSelect {...props} chip={label} />;
}
