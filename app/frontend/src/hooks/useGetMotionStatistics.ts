import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/api";

export const useGetMotionStatistics = () =>
	useQuery(orpc.motions.getStatistics.queryOptions({}));
