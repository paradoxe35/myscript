import { Search } from "lucide-react";

import { Label } from "@/components/ui/label";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarInput,
} from "@/components/ui/sidebar";

type Props = React.ComponentProps<"form"> & {
  inputProps?: React.ComponentProps<"input">;
};

export function SearchForm(props: Props) {
  const { inputProps, ...rest } = props;

  return (
    <form {...rest}>
      <SidebarGroup className="py-0">
        <SidebarGroupContent className="relative">
          <Label htmlFor="search" className="sr-only">
            Search
          </Label>
          <SidebarInput
            id="search"
            name="search"
            placeholder="Search the pages..."
            className="pl-8"
            {...inputProps}
          />
          <Search className="pointer-events-none absolute left-2 top-1/2 size-4 -translate-y-1/2 select-none opacity-50" />
        </SidebarGroupContent>
      </SidebarGroup>
    </form>
  );
}
