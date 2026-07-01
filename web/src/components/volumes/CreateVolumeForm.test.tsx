import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CreateVolumeForm } from "@/components/volumes/CreateVolumeForm";

describe("CreateVolumeForm", () => {
  it("submits no filter when extensions are empty", () => {
    const onSubmit = vi.fn();
    render(
      <CreateVolumeForm
        disks={[
          {
            path: "/data/disks/ssd",
            name: "ssd",
            label: "SSD",
            total_bytes: 1000,
            free_bytes: 500,
          },
        ]}
        onCancel={() => undefined}
        onSubmit={onSubmit}
      />,
    );

    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Docs" } });
    fireEvent.click(screen.getByRole("button", { name: "Create volume" }));

    expect(onSubmit).toHaveBeenCalledWith({
      name: "Docs",
      disk_path: "/data/disks/ssd",
      quota_bytes: 0,
      filters: { mode: "", extensions: [] },
    });
  });

  it("submits normalized values", () => {
    const onSubmit = vi.fn();
    render(
      <CreateVolumeForm
        disks={[
          {
            path: "/data/disks/ssd",
            name: "ssd",
            label: "SSD",
            total_bytes: 1000,
            free_bytes: 500,
          },
        ]}
        onCancel={() => undefined}
        onSubmit={onSubmit}
      />,
    );

    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Photos" } });
    fireEvent.change(screen.getByLabelText(/Quota/), { target: { value: "50" } });
    fireEvent.change(screen.getByPlaceholderText(".jpg .png .webp"), {
      target: { value: "jpg png" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create volume" }));

    expect(onSubmit).toHaveBeenCalledWith({
      name: "Photos",
      disk_path: "/data/disks/ssd",
      quota_bytes: 53687091200,
      filters: {
        mode: "allow",
        extensions: [".jpg", ".png"],
      },
    });
  });
});
