import {
  Activity,
  HardDrive,
  LayoutDashboard,
  ListTodo,
  Puzzle,
  Settings,
  Users,
  type LucideIcon,
} from "lucide-react";

export const navIcons: Record<string, LucideIcon> = {
  "/dashboard": LayoutDashboard,
  "/volumes": HardDrive,
  "/monitoring": Activity,
  "/plugins": Puzzle,
  "/tasks": ListTodo,
  "/settings": Settings,
  "/settings/users": Users,
};
