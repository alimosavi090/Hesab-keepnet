"use client";

import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AnimatePresence, motion } from "framer-motion";
import {
  IconLoader2,
  IconNotebook,
  IconPencil,
  IconPin,
  IconPinFilled,
  IconPlus,
  IconTrash,
} from "@tabler/icons-react";
import { notesApi } from "@/lib/api";
import type { Note, NoteEntityType } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { JalaliDate } from "@/components/shared/jalali-date";
import { SPRING } from "@/components/shared/motion";
import { cn } from "@/lib/utils";

type Props = {
  entityType: NoteEntityType;
  entityId: number;
  title: string;
};

/* Small ghost button + dialog listing notes attached to a record.
   Used on representatives, sales rows and bank account cards. */
export function EntityNotes({ entityType, entityId, title }: Props) {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        aria-label={`یادداشت‌های ${title}`}
        className={cn("text-muted-foreground transition-colors duration-300 hover:text-primary")}
        onClick={() => setOpen(true)}
      >
        <IconNotebook className="size-4" />
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="glass max-w-md">
          <DialogHeader>
            <DialogTitle>یادداشت‌ها — {title}</DialogTitle>
          </DialogHeader>
          {open ? <EntityNotesList entityType={entityType} entityId={entityId} /> : null}
        </DialogContent>
      </Dialog>
    </>
  );
}

function EntityNotesList({ entityType, entityId }: { entityType: NoteEntityType; entityId: number }) {
  const queryClient = useQueryClient();
  const [body, setBody] = useState("");
  const [editing, setEditing] = useState<Note | null>(null);

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ["notes", entityType, entityId] });

  const listQuery = useQuery({
    queryKey: ["notes", entityType, entityId],
    queryFn: () => notesApi.list({ entity_type: entityType, entity_id: entityId, page: 1, page_size: 50 }),
    enabled: true, // this list only mounts while the dialog is open
  });

  const createMutation = useMutation({
    mutationFn: (e: FormEvent) => {
      e.preventDefault();
      return notesApi.create({
        entity_type: entityType,
        entity_id: entityId,
        body: body.trim(),
      });
    },
    onSuccess: () => {
      setBody("");
      invalidate();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, text }: { id: number; text: string }) =>
      notesApi.update(id, { body: text }),
    onSuccess: () => {
      setEditing(null);
      invalidate();
    },
  });

  const pinMutation = useMutation({
    mutationFn: ({ id, pinned }: { id: number; pinned: boolean }) =>
      notesApi.update(id, { pinned }),
    onSuccess: invalidate,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => notesApi.remove(id),
    onSuccess: invalidate,
  });

  const items = [...(listQuery.data?.items ?? [])].sort(
    (a, b) => Number(b.pinned) - Number(a.pinned)
  );

  return (
    <div className="space-y-4">
      {editing ? (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            updateMutation.mutate({ id: editing.id, text: editing.body.trim() });
          }}
          className="bg-card/60 ring-primary/20 space-y-2 rounded-xl p-3 ring-1"
        >
          <Textarea
            rows={3}
            required
            autoFocus
            value={editing.body}
            onChange={(e) => setEditing({ ...editing, body: e.target.value })}
            aria-label="ویرایش یادداشت"
          />
          <div className="flex justify-end gap-2">
            <Button type="button" variant="ghost" size="sm" onClick={() => setEditing(null)}>
              انصراف
            </Button>
            <Button type="submit" size="sm" disabled={updateMutation.isPending} className="glow-primary rounded-xl">
              {updateMutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
              ذخیره
            </Button>
          </div>
        </form>
      ) : (
        <form
          onSubmit={(e) => createMutation.mutate(e)}
          className="space-y-2"
          aria-label="افزودن یادداشت"
        >
          <Textarea
            rows={2}
            required
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="یادداشتی برای این رکورد بنویسید…"
            aria-label="متن یادداشت"
          />
          <div className="flex justify-end">
            <Button type="submit" size="sm" disabled={createMutation.isPending} className="glow-primary rounded-xl">
              {createMutation.isPending ? (
                <IconLoader2 className="size-4 animate-spin" />
              ) : (
                <IconPlus className="size-4" />
              )}
              افزودن
            </Button>
          </div>
          {createMutation.isError ? (
            <p role="alert" className="text-destructive text-xs">
              {(createMutation.error as Error)?.message ?? "ثبت ناموفق بود."}
            </p>
          ) : null}
        </form>
      )}

      {listQuery.isPending ? (
        <p className="text-caption py-3 text-center">در حال دریافت…</p>
      ) : items.length === 0 ? (
        <p className="text-caption py-3 text-center">هنوز یادداشتی ثبت نشده است.</p>
      ) : (
        <ul className="max-h-64 space-y-2 overflow-y-auto pe-1">
          <AnimatePresence initial={false}>
            {items.map((note) => (
              <motion.li
                key={note.id}
                layout
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, x: -16 }}
                transition={SPRING}
                className="bg-card/60 ring-foreground/[0.05] group relative rounded-xl p-3 ring-1"
              >
                {note.pinned ? (
                  <IconPinFilled aria-hidden="true" className="text-warning absolute -top-1.5 right-2 size-3.5 bg-transparent" />
                ) : null}
                <p className="whitespace-pre-wrap ps-3 text-sm leading-6">{note.body}</p>
                <div className="mt-1.5 flex items-center justify-between">
                  <span className="text-caption numeric"><JalaliDate iso={note.created_at} /></span>
                  <span className="flex gap-0.5 opacity-0 transition-opacity duration-200 group-hover:opacity-100 focus-within:opacity-100">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-6"
                      aria-label="ویرایش یادداشت"
                      onClick={() => setEditing(note)}
                    >
                      <IconPencil className="text-muted-foreground size-3.5 transition-colors duration-300 hover:text-primary" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-6"
                      aria-label={note.pinned ? "برداشتن سنجاق" : "سنجاق کردن"}
                      onClick={() => pinMutation.mutate({ id: note.id, pinned: !note.pinned })}
                    >
                      {note.pinned ? <IconPinFilled className="text-warning size-3.5" /> : <IconPin className="size-3.5" />}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-6"
                      aria-label="حذف یادداشت"
                      onClick={() => deleteMutation.mutate(note.id)}
                    >
                      <IconTrash className="text-muted-foreground size-3.5 transition-colors duration-300 hover:text-destructive" />
                    </Button>
                  </span>
                </div>
              </motion.li>
            ))}
          </AnimatePresence>
        </ul>
      )}
    </div>
  );
}
