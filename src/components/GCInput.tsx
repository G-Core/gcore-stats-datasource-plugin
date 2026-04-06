import React, { ChangeEvent, useRef, useState } from "react";
import { LegacyForms } from "@grafana/ui";
import { debounce } from "lodash";

const { FormField } = LegacyForms;

export interface InputProps<T>
  extends Omit<React.HTMLProps<HTMLInputElement>, "value"> {
  width?: number;
  inputWidth?: number;
  labelWidth?: number;
  value?: T;
  onChange: (e: ChangeEvent<HTMLInputElement>) => void;
  tooltip?: string;
  type: string;
  label: string;
  debounce?: number;
}

export const GCInput: React.FC<InputProps<string>> = (rawProps) => {
  const { onChange, ...props } = rawProps;
  const [value, setValue] = useState(props.value);
  const debouncedFunc = useRef(
    debounce((nextValue: string) => {
      onChange({ target: { value: nextValue } } as ChangeEvent<HTMLInputElement>);
    }, props.debounce || 500)
  ).current;
  const onChangeDebounce = (e: ChangeEvent<HTMLInputElement>) => {
    const nextValue = e.target.value;
    setValue(nextValue);
    debouncedFunc(nextValue);
  };
  return <FormField {...props} value={value} onChange={onChangeDebounce} />;
};
