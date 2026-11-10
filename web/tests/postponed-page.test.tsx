import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import Postponed from "@/pages/Postponed";

const state = vi.hoisted(() => ({ listProps: [] as Array<Record<string, unknown>> }));

vi.mock("@/components/MemoView", () => ({ default: () => <div /> }));

vi.mock("@/components/PagedMemoList", () => ({
  default: (props: Record<string, unknown>) => {
    state.listProps.push(props);
    const renderLeading = props.renderLeading as ((options: { useGrid: boolean }) => React.ReactNode) | undefined;
    return <div>{renderLeading?.({ useGrid: false })}</div>;
  },
  getMemoKey: (memo: { name: string }) => memo.name,
}));

describe("Postponed page", () => {
  it("lists only hidden memos under its own heading", () => {
    render(<Postponed />);

    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("Postponed");
    expect(state.listProps.at(-1)).toMatchObject({ hiddenOnly: true, emptyMessage: "No postponed memos." });
  });
});
