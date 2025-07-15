import {
  type RouteConfig,
  route,
} from "@react-router/dev/routes";

export default [
  route("/", "./routes/_index.tsx"),
  route("/fracties", "./routes/fracties.tsx"),
  route("/moties", "./routes/moties.tsx"),
  // pattern ^           ^ module file
] satisfies RouteConfig;
