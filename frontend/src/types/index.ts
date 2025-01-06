export type NotionPage = {
  object: "page";
  id: string;
  created_time: string;
  last_edited_time: string;
  created_by: {
    object: "user";
    id: string;
  };
  last_edited_by: {
    object: "user";
    id: string;
  };
  archived: boolean;
  properties: {
    title: {
      id: string;
      type: "title";
      title: Array<{
        type: "text";
        text: {
          content: string;
        };
        annotations: {
          bold: boolean;
          italic: boolean;
          strikethrough: boolean;
          underline: boolean;
          code: boolean;
          color: string;
        };
        plain_text: string;
      }>;
    };
  };
  parent: {
    type: "workspace" | "page_id";
    workspace?: boolean;
    page_id?: string;
  };
  url: string;
  public_url: string;
};

export type NotionSimplePage = {
  id: string;
  title: string;
};

export type RepositoryBaseFields =
  | "ID"
  | "CreatedAt"
  | "UpdatedAt"
  | "DeletedAt"
  | "convertValues";

export type WithoutRepositoryBaseFields<T> = Omit<T, RepositoryBaseFields>;

export type EventClear = () => void;

export type ExtractProperties<T, K extends keyof T> = {
  [P in K]: T[P];
};
