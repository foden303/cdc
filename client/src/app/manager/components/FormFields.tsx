type InputGroupProps = {
  label: string;
  name: string;
  defaultValue?: string | number;
  placeholder?: string;
  type?: string;
  readOnly?: boolean;
  required?: boolean;
};

export function InputGroup({
  label,
  name,
  defaultValue,
  placeholder,
  type = "text",
  readOnly = false,
  required = false,
}: InputGroupProps) {
  return (
    <div className="space-y-1">
      <label className="overline-label overline-label-dim ml-0.5">
        {label}
        {required ? <span className="text-rose-500"> *</span> : null}
      </label>
      <input
        type={type}
        name={name}
        defaultValue={defaultValue}
        placeholder={placeholder}
        readOnly={readOnly}
        required={required}
        className={`w-full bg-muted/30 border border-border rounded-lg px-3 py-1.5 text-foreground outline-none focus:border-primary/40 text-base-compact font-bold transition-all placeholder:text-muted-foreground/30 ${readOnly ? "opacity-40 cursor-not-allowed" : ""}`}
      />
    </div>
  );
}

type TextAreaGroupProps = {
  label: string;
  name: string;
  defaultValue?: string;
  placeholder?: string;
  rows?: number;
  required?: boolean;
};

export function TextAreaGroup({
  label,
  name,
  defaultValue,
  placeholder,
  rows = 2,
  required = false,
}: TextAreaGroupProps) {
  return (
    <div className="space-y-1">
      <label className="overline-label overline-label-dim ml-0.5">
        {label}
        {required ? <span className="text-rose-500"> *</span> : null}
      </label>
      <textarea
        name={name}
        defaultValue={defaultValue}
        placeholder={placeholder}
        rows={rows}
        required={required}
        className="w-full bg-muted/30 border border-border rounded-lg px-3 py-1.5 text-foreground outline-none focus:border-primary/40 text-base-compact font-bold transition-all placeholder:text-muted-foreground/30 resize-y"
      />
    </div>
  );
}

const SELECT_INPUT_CLASS =
  "w-full bg-muted/30 border border-border rounded-lg px-3 py-1.5 text-foreground outline-none focus:border-primary/40 text-base-compact font-bold";

type TypeSelectProps = {
  name?: string;
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
};

export function TypeSelect({ name = "type", value, onChange, options }: TypeSelectProps) {
  return (
    <select
      name={name}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={SELECT_INPUT_CLASS}
    >
      {options.map((o) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  );
}
