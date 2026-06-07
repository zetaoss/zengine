<svelte:options
  customElement={{
    props: {
      class: { type: "String" },
      cooldown: { type: "Number" },
      disabled: { type: "Boolean" },
      href: { type: "String" },
      size: { type: "String" },
      title: { type: "String" },
      type: { type: "String" },
      variant: { type: "String" },
    },
  }}
/>

<script lang="ts" module>
// https://github.com/huntabyte/shadcn-svelte/blob/main/docs/src/lib/registry/ui/button/button.svelte

import type { Snippet } from "svelte"
import type { HTMLAnchorAttributes, HTMLButtonAttributes } from "svelte/elements"

  function cn(...classes: Array<string | false | null | undefined>): string {
    return classes.filter(Boolean).join(" ")
  }

  export type ButtonVariant = "default" | "outline" | "secondary" | "ghost" | "destructive" | "link"
  export type ButtonSize =
    | "default"
    | "xs"
    | "sm"
    | "lg"
    | "icon"
    | "icon-xs"
    | "icon-sm"
    | "icon-lg"
    | "small"
    | "medium"

  const variantClasses: Record<ButtonVariant, string> = {
    default: "cn-button-variant-default",
    outline: "cn-button-variant-outline",
    secondary: "cn-button-variant-secondary",
    ghost: "cn-button-variant-ghost",
    destructive: "cn-button-variant-destructive",
    link: "cn-button-variant-link",
  }

  const sizeClasses: Record<ButtonSize, string> = {
    default: "cn-button-size-default",
    xs: "cn-button-size-xs",
    sm: "cn-button-size-sm",
    lg: "cn-button-size-lg",
    icon: "cn-button-size-icon",
    "icon-xs": "cn-button-size-icon-xs",
    "icon-sm": "cn-button-size-icon-sm",
    "icon-lg": "cn-button-size-icon-lg",
    small: "cn-button-size-sm",
    medium: "cn-button-size-default",
  }

  export function buttonVariants(opts: { variant?: ButtonVariant; size?: ButtonSize }): string {
    return cn("cn-button", variantClasses[opts.variant ?? "outline"], sizeClasses[opts.size ?? "medium"])
  }

  type ButtonCommonProps = {
    class?: string
    ref?: HTMLAnchorElement | HTMLButtonElement | null
    variant?: ButtonVariant
    size?: ButtonSize
    cooldown?: number
    disabled?: boolean
    title?: string
    type?: HTMLButtonAttributes["type"]
    onclick?: ((event: MouseEvent) => void) | undefined
    children?: Snippet
  }

  type ButtonAnchorProps = ButtonCommonProps &
    Omit<HTMLAnchorAttributes, "class" | "href" | "type"> & {
      href: string
    }

  type ButtonNativeProps = ButtonCommonProps &
    Omit<HTMLButtonAttributes, "class" | "type"> & {
      href?: undefined
    }

  export type ButtonProps = ButtonAnchorProps | ButtonNativeProps
</script>

<script lang="ts">
  let {
    class: className,
    variant = "outline",
    size = "medium",
    ref = $bindable(null),
    href = undefined,
    type = "button",
    disabled,
    cooldown = 300,
    title = undefined,
    onclick,
    children,
    ...restProps
  }: ButtonProps = $props()

  let isCooling = $state(false)

  let isDisabled = $derived(disabled || isCooling)
  let isLink = $derived(href !== undefined)
  let classes = $derived(buttonVariants({ variant, size }))

  function handleClick(event: MouseEvent) {
    if (isDisabled) {
      event.preventDefault()
      event.stopImmediatePropagation()
      return
    }

    onclick?.(event)

    // Keep native navigation intact for anchor buttons.
    if (!isLink && cooldown > 0) {
      isCooling = true
      setTimeout(() => {
        isCooling = false
      }, cooldown)
    }
  }
</script>

{#if href}
  <a
    bind:this={ref}
    data-slot="button"
    {...(restProps as Omit<HTMLAnchorAttributes, "class" | "href" | "type">)}
    href={isDisabled ? undefined : href}
    aria-label={title}
    class={cn(classes, className ?? "", isDisabled ? "pointer-events-none opacity-50" : "")}
    aria-disabled={isDisabled ? "true" : undefined}
    role={isDisabled ? "link" : undefined}
    tabindex={isDisabled ? -1 : undefined}
    onclick={handleClick}
  >
    {@render children?.()}
  </a>
{:else}
  <button
    bind:this={ref}
    data-slot="button"
    {...(restProps as Omit<HTMLButtonAttributes, "class" | "type">)}
    {type}
    aria-label={title}
    class={cn(classes, className ?? "")}
    disabled={isDisabled}
    onclick={handleClick}
  >
    {@render children?.()}
  </button>
{/if}
