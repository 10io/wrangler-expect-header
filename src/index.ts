export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    url.host = '0.0.0.0:8080';

    const newRequest = new Request(url.toString(), request);
    return fetch(newRequest);
  },
} satisfies ExportedHandler<Env>;
