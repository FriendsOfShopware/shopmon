<template>
  <div :class="rootClasses">
    <label v-if="label" :id="labelId" :for="nativeSelectId">{{ label }}</label>

    <select
      :id="nativeSelectId"
      :name="name"
      class="sr-only"
      :class="{ 'has-error': error }"
      :value="stringValue"
      aria-hidden="true"
      tabindex="-1"
      v-bind="nativeSelectAttrs"
      @change="onNativeChange"
    >
      <slot />
    </select>

    <Listbox :model-value="stringValue" :disabled="isDisabled" @update:model-value="selectValue">
      <div class="select-shell">
        <ListboxButton
          :id="buttonId"
          class="field select-trigger"
          :class="{ 'has-error': error }"
          :aria-labelledby="label ? labelId : undefined"
          :aria-invalid="error ? 'true' : undefined"
          v-bind="triggerAttrs"
        >
          <span
            class="select-trigger-label"
            :class="{ 'select-trigger-placeholder': !selectedOption }"
          >
            {{ selectedOption?.label ?? placeholderText }}
          </span>

          <icon-fa6-solid:chevron-down class="select-trigger-icon" />
        </ListboxButton>

        <transition
          enter-active-class="transition ease-out duration-100"
          enter-from-class="transform opacity-0 scale-95"
          enter-to-class="transform opacity-100 scale-100"
          leave-active-class="transition ease-in duration-75"
          leave-from-class="transform opacity-100 scale-100"
          leave-to-class="transform opacity-0 scale-95"
        >
          <ListboxOptions class="select-options">
            <ListboxOption
              v-for="option in parsedOptions"
              :key="option.key"
              :value="option.value"
              :disabled="option.disabled"
              as="template"
              v-slot="{ active, selected, disabled }"
            >
              <li
                class="select-option"
                :class="{
                  'is-active': active,
                  'is-selected': selected,
                  'is-disabled': disabled,
                }"
              >
                <span class="select-option-label">{{ option.label }}</span>
                <icon-fa6-solid:check v-if="selected" class="select-option-indicator" />
              </li>
            </ListboxOption>
          </ListboxOptions>
        </transition>
      </div>
    </Listbox>

    <div v-if="error" class="field-error-message">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import {
  Comment,
  Fragment,
  Text,
  computed,
  isVNode,
  useAttrs,
  useId,
  useSlots,
  type VNode,
  type VNodeArrayChildren,
} from "vue";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";

defineOptions({ inheritAttrs: false });

interface ParsedOption {
  key: string;
  value: string;
  label: string;
  disabled: boolean;
}

const props = withDefaults(
  defineProps<{
    name?: string;
    id?: string;
    label?: string;
    error?: string;
    modelValue?: string | number;
    placeholder?: string;
  }>(),
  {
    name: undefined,
    id: undefined,
    label: undefined,
    error: undefined,
    modelValue: undefined,
    placeholder: "Select option",
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
  change: [value: string];
}>();

const attrs = useAttrs();
const slots = useSlots();
const generatedId = useId();

const baseId = computed(() => props.id ?? props.name ?? `select-${generatedId}`);
const nativeSelectId = computed(() => baseId.value);
const buttonId = computed(() => `${baseId.value}-trigger`);
const labelId = computed(() => `${baseId.value}-label`);

const stringValue = computed(() => {
  return props.modelValue === undefined || props.modelValue === null ? "" : String(props.modelValue);
});

const isDisabled = computed(() => {
  return attrs.disabled !== undefined && attrs.disabled !== false;
});

const placeholderText = computed(() => {
  return props.placeholder;
});

const rootClasses = computed(() => {
  return attrs.class;
});

const nativeSelectAttrs = computed(() => {
  const values = { ...attrs };
  delete values.class;
  delete values.style;
  return values;
});

const triggerAttrs = computed(() => {
  const values = { ...attrs };
  delete values.class;
  delete values.style;
  delete values.name;
  delete values.required;
  delete values.form;
  delete values.multiple;
  return values;
});

const parsedOptions = computed<ParsedOption[]>(() => {
  const nodes = slots.default?.() ?? [];
  return extractOptions(nodes);
});

const selectedOption = computed(() => {
  return parsedOptions.value.find((option) => option.value === stringValue.value);
});

function selectValue(value: string) {
  emit("update:modelValue", value);
  emit("change", value);
}

function onNativeChange(event: Event) {
  selectValue((event.target as HTMLSelectElement).value);
}

function extractOptions(nodes: VNodeArrayChildren, options: ParsedOption[] = []): ParsedOption[] {
  for (const node of nodes) {
    if (Array.isArray(node)) {
      extractOptions(node, options);
      continue;
    }

    if (typeof node === "string" || typeof node === "number") {
      continue;
    }

    if (!isVNode(node) || node.type === Comment || node.type === Text) {
      continue;
    }

    if (node.type === Fragment) {
      extractOptions((node.children as VNodeArrayChildren) ?? [], options);
      continue;
    }

    if (node.type !== "option") {
      continue;
    }

    const nodeProps = (node.props ?? {}) as Record<string, unknown>;
    const label = extractText(node).trim();
    const rawValue = nodeProps.value ?? label;

    options.push({
      key: String(node.key ?? `${rawValue}-${options.length}`),
      value: String(rawValue ?? ""),
      label,
      disabled: nodeProps.disabled !== undefined && nodeProps.disabled !== false,
    });
  }

  return options;
}

function extractText(node: VNode): string {
  const children = node.children;

  if (typeof children === "string") {
    return children;
  }

  if (Array.isArray(children)) {
    return children.map((child) => extractChildText(child)).join("");
  }

  return "";
}

function extractChildText(child: VNodeArrayChildren[number]): string {
  if (typeof child === "string" || typeof child === "number") {
    return String(child);
  }

  if (Array.isArray(child)) {
    return child.map((nestedChild) => extractChildText(nestedChild)).join("");
  }

  if (!isVNode(child)) {
    return "";
  }

  if (child.type === Text || child.type === Comment) {
    return typeof child.children === "string" ? child.children : "";
  }

  if (child.type === Fragment) {
    return Array.isArray(child.children)
      ? child.children.map((nestedChild) => extractChildText(nestedChild)).join("")
      : "";
  }

  return typeof child.children === "string" ? child.children : "";
}
</script>

<style>
.select-shell {
  position: relative;
}

.select-trigger {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
  font-weight: 400;
  padding-right: 2.25rem;
}

.select-trigger-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-trigger-placeholder {
  color: var(--field-placeholder-color);
}

.select-trigger-icon {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  width: 1rem;
  height: 1rem;
  transform: translateY(-50%);
  color: var(--text-color-muted);
  pointer-events: none;
}

.select-options {
  position: absolute;
  z-index: 30;
  margin-top: 0.25rem;
  min-width: calc(100% + 3px);
  width: max-content;
  max-width: min(100%, 28rem);
  max-height: 24rem;
  overflow-y: auto;
  padding: 0.375rem;
  background-color: var(--panel-background);
  color: var(--text-color);
  border-radius: 0.5rem;
  box-shadow:
    inset 0 0 0 1px var(--field-border-color),
    var(--surface-shadow-strong);
}

.select-option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 1rem;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 1rem;
  cursor: pointer;
  list-style: none;

  &.is-active {
    background-color: var(--item-hover-background);
  }

  &.is-disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.select-option-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-option-indicator {
  width: 1rem;
  height: 1rem;
  justify-self: end;
}
</style>
