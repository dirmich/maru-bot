// Client-side DB mock
// The actual database operations are performed by the Go backend via API
export const db = new Proxy({}, {
  get: function (target, prop) {
    return function () {
      console.warn(`Database operation '${String(prop)}' called on client-side. This should be an API call.`);
      return Promise.resolve([]);
    }
  }
});
