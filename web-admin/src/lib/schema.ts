// Client-side schema mock
// Schema types should be shared or generated, but runtime schema code (drizzle) is not needed in client bundle
export type User = {
    id: string;
    name?: string;
    email?: string;
    image?: string;
};

export type AdapterAccount = any;
