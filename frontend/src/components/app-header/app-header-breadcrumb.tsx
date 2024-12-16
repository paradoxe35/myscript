import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
} from "@/components/ui/breadcrumb";
import { useActivePageStore } from "@/store/active-page";

export function AppHeaderBreadcrumb() {
  const activePage = useActivePageStore((state) => state.page);

  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem className="flex items-center justify-between">
          <BreadcrumbLink href="#">{activePage?.page.title}</BreadcrumbLink>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  );
}
