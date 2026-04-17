"use client";

import type { FormEvent } from "react";
import { Database, Box, X, Save, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/Button";
import type { TranslationKey } from "@/lib/i18n";
import { buildConfigPayloadFromFormData, mapToText } from "../utils/form";
import {
  SOURCE_TYPE_OPTIONS,
  SINK_TYPE_OPTIONS,
  getSourcePortPlaceholder,
  sinkShowsElasticsearchFields,
  sinkShowsRedisFields,
} from "../registry";
import { InputGroup, TextAreaGroup, TypeSelect } from "./FormFields";

type TFn = (key: TranslationKey) => string;

export type SourceSinkModalKind = "source" | "sink";

type Props = {
  kind: SourceSinkModalKind;
  editingItem: Record<string, any> | null;
  selectedType: string;
  onSelectedTypeChange: (type: string) => void;
  isSaving: boolean;
  onClose: () => void;
  onSave: (payload: Record<string, unknown>) => void | Promise<void>;
  t: TFn;
};

export function SourceSinkModal({
  kind,
  editingItem,
  selectedType,
  onSelectedTypeChange,
  isSaving,
  onClose,
  onSave,
  t,
}: Props) {
  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const payload = buildConfigPayloadFromFormData(formData, kind === "sink" ? selectedType : "");
    await onSave(payload);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="w-full max-w-lg bg-card border border-border rounded-2xl shadow-2xl overflow-hidden shadow-primary/5">
        <div className="flex items-center justify-between p-4 border-b border-border bg-muted/30">
          <div className="flex items-center gap-3">
            <div
              className={`w-8 h-8 rounded-lg flex items-center justify-center ${
                kind === "source" ? "bg-blue-500/10 text-blue-500" : "bg-emerald-500/10 text-emerald-500"
              }`}
            >
              {kind === "source" ? <Database className="w-4 h-4" /> : <Box className="w-4 h-4" />}
            </div>
            <div>
              <h3 className="text-sm font-black">
                {editingItem ? "Edit" : "Add"} {kind === "source" ? t("ingestionSources") : t("deliverySinks")}
              </h3>
            </div>
          </div>
          <button type="button" onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-3.5 max-h-[80vh] overflow-y-auto custom-scrollbar">
          <div className="grid grid-cols-2 gap-3">
            <InputGroup label="Name" name="name" defaultValue={editingItem?.name} placeholder="Production DB" />
            <InputGroup
              label="ID"
              name="instance_id"
              defaultValue={editingItem?.instance_id}
              placeholder="prod-01"
              readOnly={!!editingItem}
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="overline-label overline-label-dim ml-0.5">
                Type <span className="text-rose-500">*</span>
              </label>
              <TypeSelect
                value={selectedType}
                onChange={onSelectedTypeChange}
                options={kind === "source" ? SOURCE_TYPE_OPTIONS : SINK_TYPE_OPTIONS}
              />
            </div>
            <InputGroup
              label="Topic"
              name="topic"
              defaultValue={editingItem?.topic}
              placeholder={kind === "source" ? "cdc.orders" : "sink.orders"}
            />
          </div>

          {kind === "source" ? (
            <>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Host" name="host" defaultValue={editingItem?.host} placeholder="localhost" required />
                <InputGroup
                  label="Port"
                  name="port"
                  type="number"
                  defaultValue={editingItem?.port}
                  placeholder={getSourcePortPlaceholder(selectedType)}
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Database" name="database" defaultValue={editingItem?.database} placeholder="app_db" required />
                <TextAreaGroup
                  label="Tables (comma)"
                  name="tables"
                  defaultValue={editingItem?.tables?.join(", ")}
                  placeholder="orders, users, invoices"
                  rows={2}
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Username" name="username" defaultValue={editingItem?.username} placeholder="cdc_user" />
                <InputGroup label="Password" name="password" type="password" defaultValue={editingItem?.password} placeholder="••••••••" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Slot Name" name="slot_name" defaultValue={editingItem?.slot_name} placeholder="cdc_slot" />
                <InputGroup
                  label="Publication"
                  name="publication_name"
                  defaultValue={editingItem?.publication_name}
                  placeholder="cdc_publication"
                />
              </div>
            </>
          ) : (
            <>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="URL" name="url" defaultValue={editingItem?.url?.[0]} placeholder="http://localhost:8123" required />
                <InputGroup label="Batch Size" name="batch_size" type="number" defaultValue={editingItem?.batch_size} placeholder="500" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Flush Interval (ms)" name="flush_interval_ms" type="number" defaultValue={editingItem?.flush_interval_ms} placeholder="1000" />
                <InputGroup label="Max Retries" name="max_retries" type="number" defaultValue={editingItem?.max_retries} placeholder="5" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Retry Base (ms)" name="retry_base_ms" type="number" defaultValue={editingItem?.retry_base_ms} placeholder="250" />
                <InputGroup label="API Key" name="api_key" defaultValue={editingItem?.api_key} placeholder="token..." />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <InputGroup label="Username" name="username" defaultValue={editingItem?.username} placeholder="sink_user" />
                <InputGroup label="Password" name="password" type="password" defaultValue={editingItem?.password} placeholder="••••••••" />
              </div>

              {sinkShowsElasticsearchFields(selectedType) && (
                <>
                  <div className="grid grid-cols-2 gap-3">
                    <InputGroup label="Index Prefix" name="index_prefix" defaultValue={editingItem?.index_prefix} placeholder="cdc-" />
                    <InputGroup label="Index" name="index" defaultValue={editingItem?.index} placeholder="orders-v1" />
                  </div>
                  <TextAreaGroup
                    label="Index Mapping (src:dst, ...)"
                    name="index_mapping"
                    defaultValue={mapToText(editingItem?.index_mapping)}
                    placeholder="created_at:@timestamp, id:_id"
                    rows={2}
                  />
                </>
              )}

              <TextAreaGroup
                label="Field Mapping (src:dst, ...)"
                name="field_mapping"
                defaultValue={mapToText(editingItem?.field_mapping)}
                placeholder="order_id:id, customer_email:email"
                rows={2}
              />

              {sinkShowsRedisFields(selectedType) && (
                <>
                  <div className="grid grid-cols-2 gap-3">
                    <InputGroup label="Redis Command" name="redis_command" defaultValue={editingItem?.redis?.command} placeholder="set" />
                    <InputGroup label="Key Template" name="redis_key_template" defaultValue={editingItem?.redis?.key_template} placeholder="{topic}:{id}" />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <TextAreaGroup
                      label="Value Fields (comma)"
                      name="redis_value_fields"
                      defaultValue={editingItem?.redis?.value_fields?.join(", ")}
                      placeholder="id, status, updated_at"
                      rows={2}
                    />
                    <InputGroup label="TTL (seconds)" name="redis_ttl" type="number" defaultValue={editingItem?.redis?.ttl} placeholder="3600" />
                  </div>
                </>
              )}
            </>
          )}

          <div className="pt-4 border-t border-border flex gap-3">
            <Button type="submit" disabled={isSaving} className="flex-1">
              {isSaving ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Save className="w-3.5 h-3.5" />}
              {editingItem ? t("save") : t("add")}
            </Button>
            <Button variant="muted" type="button" onClick={onClose}>
              {t("cancel")}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
