import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";

type BreadcrumbNavProps = {
  volumeId: string;
  path: string;
  onNavigate: (path: string) => void;
};

function splitPath(path: string): string[] {
  if (path === "." || path === "") {
    return [];
  }
  return path.split("/").filter(Boolean);
}

export function BreadcrumbNav({ volumeId, path, onNavigate }: BreadcrumbNavProps) {
  const segments = splitPath(path);

  return (
    <nav className="flex flex-wrap items-center gap-2 text-sm text-body">
      <Link to={`/volumes/${volumeId}`} className="text-primary hover:underline">
        root
      </Link>
      {segments.map((segment, index) => {
        const target = segments.slice(0, index + 1).join("/");
        const isLast = index === segments.length - 1;
        return (
          <span key={target} className="flex items-center gap-2">
            <span className="text-muted-soft">/</span>
            <button
              type="button"
              className={cn(
                "hover:underline",
                isLast ? "font-medium text-ink" : "text-primary",
              )}
              onClick={() => onNavigate(target)}
            >
              {segment}
            </button>
          </span>
        );
      })}
    </nav>
  );
}
