"use client";

import { useMemo, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { IconLoader2, IconPencilPlus, IconPin, IconPinFilled, IconPlus, IconSearch, IconTrash } from "@tabler/icons-react";
import { notesApi } from "@/lib/api";
import type { Note } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { PageToolbar, ToolbarSpacer } from "@/components/shared/page-toolbar";
import { JalaliDate } from "@/components/shared/jalali-date";
import { SPRING, Stagger, StaggerItem } from "@/components/shared/motion";

export function parseNoteTags(tags: string): string[] {
  return tags.split(",").map((t) => t.trim()).filter(Boolean);
}

export default function JournalPage() {
  const queryClient = useQueryClient();
  const [query, setQuery] = useState("");
  const [tagFilter, setTagFilter] = useState<string>("all");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<Note | null>(null);

  const notesQuery = useQuery({
    queryKey: ["notes", "journal", query],
    queryFn: () =>
      notesApi.list({ entity_type: "JOURNAL", page: 1, page_size: 200, q: query || undefined }),
  });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["notes"] });

  const pinMutation = useMutation({
    mutationFn: ({ id, pinned }: { id: number; pinned: boolean }) =>
      notesApi.update(id, { pinned }),
    onSuccess: invalidate,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => notesApi.remove(id),
    onSuccess: invalidate,
  });

  const items = useMemo(() => notesQuery.data?.items ?? [], [notesQuery.data]);
  const allTags = useMemo(() => {
    const s = new Set<string>();
    for (const n of items) for (const t of parseNoteTags(n.tags)) s.add(t);
    return [...s].sort();
  }, [items]);

  const visible = items.filter(
    (n) => tagFilter === "all" || parseNoteTags(n.tags).includes(tagFilter)
  );
  const pinnedFirst = [
    ...visible.filter((n) => n.pinned),
    ...visible.filter((n) => !n.pinned),
  ];

  return (
    <div className="mx-auto w-full max-w-3xl space-y-5">
      <PageToolbar>
        <div className="relative flex-1 min-w-40">
          <IconSearch
            aria-hidden="true"
            className="text-muted-foreground absolute right-3 top-1/2 size-4 -translate-y-1/2"
          />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="جستجو در دفتر…"
            aria-label="جستجو در دفتر روزانه"
            className="rounded-full pr-9"
          />
        </div>
        <ToolbarSpacer />
        <Button onClick={() => { setEditing(null); setEditorOpen(true); }} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          یادداشت جدید
        </Button>
      </PageToolbar>

      {allTags.length > 0 ? (
        <div className="flex flex-wrap gap-1.5 px-1" role="group" aria-label="فیلتر برچسب‌ها">
          <button
            type="button"
            aria-pressed={tagFilter === "all"}
            onClick={() => setTagFilter("all")}
            className={`border-border/60 hover:bg-accent rounded-full border px-3 py-1 text-xs transition-colors ${
              tagFilter === "all" ? "bg-primary text-primary-foreground glow-primary border-transparent font-bold" : ""
            }`}
          >
            همه
          </button>
          {allTags.map((tag) => (
            <button
              key={tag}
              type="button"
              aria-pressed={tagFilter === tag}
              onClick={() => setTagFilter(tag)}
              className={`border-border/60 hover:bg-accent rounded-full border px-3 py-1 text-xs transition-colors ${
                tagFilter === tag ? "bg-primary text-primary-foreground glow-primary border-transparent font-bold" : ""
              }`}
            >
              #{tag}
            </button>
          ))}
        </div>
      ) : null}

      {notesQuery.isPending ? <LoadingState /> : null}
      {notesQuery.isError ? <ErrorState onRetry={() => notesQuery.refetch()} /> : null}

      {notesQuery.data && pinnedFirst.length === 0 ? (
        <EmptyState
          icon={IconPencilPlus}
          title={query || tagFilter !== "all" ? "یادداشتی مطابق جستجو نیست" : "دفتر روزانه خالی است"}
          description="اولین یادداشت خود را ثبت کنید؛ ایده‌ها، جمع‌بندی روز یا نکات مالی."
          action={
            <Button size="sm" onClick={() => { setEditing(null); setEditorOpen(true); }}>
              <IconPlus className="size-4" />
              نوشتن یادداشت
            </Button>
          }
        />
      ) : null}

      <Stagger step={0.05} className="space-y-3">
        {pinnedFirst.map((note) => (
          <StaggerItem key={note.id}>
            <JournalNoteCard
              note={note}
              onEdit={() => { setEditing(note); setEditorOpen(true); }}
              onTogglePin={() =>
                pinMutation.mutate({ id: note.id, pinned: !note.pinned })
              }
              onDelete={() => deleteMutation.mutate(note.id)}
            />
          </StaggerItem>
        ))}
      </Stagger>

      <JournalEditorDialog
        open={editorOpen}
        editing={editing}
        onClose={() => setEditorOpen(false)}
        onSaved={invalidate}
      />
    </div>
  );
}

function JournalNoteCard({
  note,
  onEdit,
  onTogglePin,
  onDelete,
}: {
  note: Note;
  onEdit: () => void;
  onTogglePin: () => void;
  onDelete: () => void;
}) {
  return (
    <motion.article
      layout
      whileHover={{ y: -2 }}
      transition={SPRING}
      className={`glass sheen ring-foreground/[0.06] group relative rounded-2xl p-4 shadow-sm ${
        note.pinned ? "ring-warning/30" : ""
      }`}
    >
      <p className="whitespace-pre-wrap leading-7">{note.body}</p>

      <div className="mt-3 flex flex-wrap items-center gap-1.5">
        {parseNoteTags(note.tags).map((t) => (
          <Badge key={t} variant="outline" className="h-5 px-2 text-[10px]">#{t}</Badge>
        ))}
        <span className="text-caption numeric ms-auto">
          <JalaliDate iso={note.created_at} />
        </span>
      </div>

      <div className="absolute left-3 top-3 flex gap-0.5 opacity-0 transition-opacity duration-300 group-hover:opacity-100 focus-within:opacity-100">
        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          aria-label={note.pinned ? "برداشتن سنجاق" : "سنجاق کردن"}
          onClick={onTogglePin}
        >
          {note.pinned ? (
            <IconPinFilled className="text-warning size-4" />
          ) : (
            <IconPin className="size-4" />
          )}
        </Button>
        <Button variant="ghost" size="icon" className="size-7" aria-label="ویرایش" onClick={onEdit}>
          <IconPencilPlus className="size-4" />
        </Button>
        <Button variant="ghost" size="icon" className="size-7" aria-label="حذف" onClick={onDelete}>
          <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
        </Button>
      </div>
    </motion.article>
  );
}

function JournalEditorDialog({
  open,
  editing,
  onClose,
  onSaved,
}: {
  open: boolean;
  editing: Note | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="glass max-w-lg">
        <DialogHeader>
          <DialogTitle>{editing ? "ویرایش یادداشت" : "یادداشت جدید"}</DialogTitle>
        </DialogHeader>
        {open ? (
          <JournalEditorForm
            key={editing?.id ?? "new"}
            editing={editing}
            onClose={onClose}
            onSaved={onSaved}
          />
        ) : null}
      </DialogContent>
    </Dialog>
  );
}

function JournalEditorForm({
  editing,
  onClose,
  onSaved,
}: {
  editing: Note | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [body, setBody] = useState(editing?.body ?? "");
  const [tags, setTags] = useState(parseNoteTags(editing?.tags ?? "").join("، "));
  const [pinned, setPinned] = useState(editing?.pinned ?? false);

  const saveMutation = useMutation({
    mutationFn: (e: FormEvent) => {
      e.preventDefault();
      const payload = {
        body: body.trim(),
        tags: tags.split(/[،,]/).map((t) => t.trim()).filter(Boolean),
        pinned,
      };
      return editing?.id
        ? notesApi.update(editing.id, payload)
        : notesApi.create({ entity_type: "JOURNAL", ...payload });
    },
    onSuccess: () => {
      onSaved();
      onClose();
    },
  });

  return (
    <form onSubmit={(e) => saveMutation.mutate(e)} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="journal-body">متن</Label>
        <Textarea
          id="journal-body"
          rows={5}
          required
          autoFocus
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="امروز چه اتفاقی افتاد؟"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="journal-tags">برچسب‌ها (با ویرگول جدا کنید)</Label>
        <Input
          id="journal-tags"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
          placeholder="مثلاً کار، ایده"
        />
      </div>
      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={pinned}
          onChange={(e) => setPinned(e.target.checked)}
          className="accent-primary size-4"
        />
        سنجاق به بالای دفتر
      </label>

      {saveMutation.isError ? (
        <p role="alert" className="text-destructive text-sm">
          {(saveMutation.error as Error)?.message ?? "ذخیره ناموفق بود."}
        </p>
      ) : null}

      <DialogFooter>
        <Button type="submit" disabled={saveMutation.isPending} className="glow-primary rounded-xl">
          {saveMutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
          ذخیره
        </Button>
      </DialogFooter>
    </form>
  );
}
