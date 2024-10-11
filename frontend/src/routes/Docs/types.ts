export interface Document {
  data: {
    title?: string;
    description?: string;
  };
  content: string;
}

export type SidebarItem =
  | {
      title: string;
      items: SidebarItem[];
    }
  | {
      title: string;
      slug: string;
      path: string;
      items?: never;
    };
