import * as React from "react";
import { cn } from "../../lib/utils";

export interface ProgressProps extends React.HTMLAttributes<HTMLDivElement> {
	value?: number;
	max?: number;
	showValue?: boolean;
	size?: "sm" | "md" | "lg";
	variant?: "default" | "success" | "warning" | "error";
}

const Progress = React.forwardRef<HTMLDivElement, ProgressProps>(
	(
		{
			className,
			value = 0,
			max = 100,
			showValue = false,
			size = "md",
			variant = "default",
			...props
		},
		ref,
	) => {
		const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

		return (
			<div
				ref={ref}
				className={cn("w-full", className)}
				role="progressbar"
				aria-valuemin={0}
				aria-valuemax={max}
				aria-valuenow={value}
				{...props}
			>
				{showValue && (
					<div className="flex justify-between text-sm text-gray-600 mb-1">
						<span>Voortgang</span>
						<span>{Math.round(percentage)}%</span>
					</div>
				)}
				<div
					className={cn("w-full bg-gray-200 rounded-full overflow-hidden", {
						"h-2": size === "sm",
						"h-3": size === "md",
						"h-4": size === "lg",
					})}
				>
					<div
						className={cn(
							"h-full transition-all duration-300 ease-in-out rounded-full",
							{
								"bg-primary-600": variant === "default",
								"bg-green-600": variant === "success",
								"bg-yellow-600": variant === "warning",
								"bg-red-600": variant === "error",
							},
						)}
						style={{ width: `${percentage}%` }}
					/>
				</div>
			</div>
		);
	},
);

Progress.displayName = "Progress";

export { Progress };
