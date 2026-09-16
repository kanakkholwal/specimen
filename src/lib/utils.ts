import { type ClassValue, clsx } from 'clsx';
import { extendTailwindMerge } from 'tailwind-merge';

// Without these, tailwind-merge reads `text-body` as a colour and drops it next to
// `text-foreground`. Register every new role size here.
const twMerge = extendTailwindMerge({
	extend: {
		classGroups: {
			'font-size': [
				{
					text: [
						'caption',
						'body',
						'body-lg',
						'body-xl',
						'subheading',
						'heading-sm',
						'heading',
						'heading-lg',
						'display',
						'display-xl'
					]
				}
			]
		}
	}
});

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// Type helpers the shadcn-svelte primitives import from here.
export type WithoutChild<T> = T extends { child?: unknown } ? Omit<T, 'child'> : T;

export type WithoutChildren<T> = T extends { children?: unknown } ? Omit<T, 'children'> : T;

export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;

export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & {
	ref?: U | null;
};
