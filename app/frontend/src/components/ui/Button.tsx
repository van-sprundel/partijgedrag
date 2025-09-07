import * as React from "react";
import { cn } from "../../lib/utils";

export interface ButtonProps
	extends React.ButtonHTMLAttributes<HTMLButtonElement> {
	variant?: "primary" | "secondary" | "ghost" | "destructive";
	size?: "sm" | "md" | "lg";
	loading?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
	(
		{
			className,
			variant = "primary",
			size = "md",
			loading = false,
			disabled,
			children,
			...props
		},
		ref,
	) => {
		return (
			<button
				className={cn(
					"btn focus-ring",
					{
						"btn-primary": variant === "primary",
						"btn-secondary": variant === "secondary",
						"btn-ghost": variant === "ghost",
						"bg-red-600 text-white hover:bg-red-700": variant === "destructive",
						"btn-sm": size === "sm",
						"btn-md": size === "md",
						"btn-lg": size === "lg",
					},
					className,
				)}
				disabled={disabled || loading}
				ref={ref}
				{...props}
			>
				{loading && (
					<svg
						className="animate-spin -ml-1 mr-2 h-4 w-4 text-current"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
					>
						<circle
							className="opacity-25"
							cx="12"
							cy="12"
							r="10"
							stroke="currentColor"
							strokeWidth="4"
						></circle>
						<path
							className="opacity-75"
							fill="currentColor"
							d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
						></path>
					</svg>
				)}
				{children}
			</button>
		);
	},
);
Button.displayName = "Button";

export { Button };
