import { Button } from "@/components/Button";
import { format } from "timeago.js";
import { Link } from "react-router";

interface RegistryStats {
	totalCount: number;
	recentUpdates: Array<{ 
		name: string; 
		version?: string; 
		published: string;
		namespace: string;
		itemName: string;
		target?: string;
	}>;
}

interface ListingStatsSidebarProps {
	type: "providers" | "modules";
	stats: RegistryStats;
	addButtonHref: string;
}

export function ListingStatsSidebar({
	type,
	stats,
	addButtonHref,
}: ListingStatsSidebarProps) {
	const formatDate = (dateString: string) => {
		return format(dateString);
	};

	return (
		<div className="p-4 space-y-6">
			<div>
				<h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">
					Registry Stats
				</h3>
				<div className="text-2xl font-bold text-brand-600 dark:text-brand-400 mb-1">
					{stats.totalCount.toLocaleString()}
				</div>
				<div className="text-sm text-gray-600 dark:text-gray-400">
					{type === "providers" ? "Providers" : "Modules"}
				</div>
			</div>

			<div>
				<Button
					href={addButtonHref}
					target="_blank"
					rel="noopener noreferrer"
					variant="primary"
					className="w-full justify-center"
				>
					Add {type === "providers" ? "Provider" : "Module"}
				</Button>
			</div>

			{stats.recentUpdates.length > 0 && (
				<div>
					<h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">
						Recent Updates
					</h3>
					<ul className="space-y-2">
						{stats.recentUpdates.map((item, index) => {
							const href = type === "modules" && item.target
								? `/module/${item.namespace}/${item.itemName}/${item.target}/latest`
								: `/provider/${item.namespace}/${item.itemName}/latest`;
							
							return (
								<li key={index} className="text-sm">
									<Link to={href} className="block hover:bg-gray-50 dark:hover:bg-gray-800 p-1 -m-1 rounded">
										<div className="flex items-center gap-2">
											<img
												src={`https://avatars.githubusercontent.com/${item.namespace}`}
												alt={`${item.namespace} avatar`}
												className="h-6 w-6 rounded ring-1 ring-gray-200 dark:ring-gray-700 flex-shrink-0"
												loading="lazy"
												onError={(e) => {
													e.currentTarget.src = "/favicon.ico";
												}}
											/>
											<div className="min-w-0 flex-grow">
												<div className="text-gray-700 dark:text-gray-300 truncate">
													{item.name}
													{item.version && (
														<span className="text-brand-600 dark:text-brand-400 ml-1">
															{item.version}
														</span>
													)}
												</div>
												<div className="text-xs text-gray-500 dark:text-gray-400">
													{formatDate(item.published)}
												</div>
											</div>
										</div>
									</Link>
								</li>
							);
						})}
					</ul>
				</div>
			)}
		</div>
	);
}
