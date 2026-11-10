import { AlarmClockIcon } from "lucide-react";
import MemoView from "@/components/MemoView";
import PagedMemoList, { getMemoKey } from "@/components/PagedMemoList";
import { cn } from "@/lib/utils";
import { Memo } from "@/types/proto/api/v1/memo_service_pb";
import { useTranslate } from "@/utils/i18n";

const Postponed = () => {
  const t = useTranslate();

  return (
    <PagedMemoList
      renderer={(memo: Memo, { compact }) => <MemoView key={getMemoKey(memo)} memo={memo} showVisibility showSpace compact={compact} />}
      hiddenOnly
      orderBy="create_time desc"
      emptyMessage={t("memo.postponed.empty")}
      renderLeading={({ useGrid }) => (
        <header className={cn("flex items-center gap-2 px-1", !useGrid && "mb-4")}>
          <AlarmClockIcon className="size-5 text-muted-foreground" strokeWidth={1.8} />
          <h1 className="text-xl font-semibold tracking-tight text-foreground">{t("memo.postponed.title")}</h1>
        </header>
      )}
    />
  );
};

export default Postponed;
