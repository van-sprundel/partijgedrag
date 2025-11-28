import express from "express";

export function handleUriError(
  err: Error,
  _req: express.Request,
  res: express.Response,
  next: express.NextFunction,
) {
  if (err instanceof URIError) {
    console.warn(`Malformed URI detected: ${err.message}`);
    res.status(400).json({
      error: "Bad Request",
      message: "Malformed URI",
    });
    return;
  }

  next(err);
}
