import { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import {
  ChevronRight,
  ChevronLeft,
  ArrowRight,
  AlertTriangle,
  Play,
  HelpCircle,
  FolderSync,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  useConfig,
  useCreateFlow,
  useDiscoverSourceTables,
  useDiscoverSinkTables,
} from '@/lib/query/manager';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { isTypeCompatible } from '@/lib/typecompat';
import type { ColumnMapping } from '@/types/api';

interface FlowWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function FlowWizard({ open, onOpenChange }: FlowWizardProps) {
  const { t } = useTranslation();
  const [step, setStep] = useState(1);

  // Form State
  const [flowName, setFlowName] = useState('');
  const [selectedSourceId, setSelectedSourceId] = useState('');
  const [selectedSinkId, setSelectedSinkId] = useState('');
  const [selectedSourceTable, setSelectedSourceTable] = useState('');
  const [selectedSinkTable, setSelectedSinkTable] = useState('');
  const [columnMappings, setColumnMappings] = useState<ColumnMapping[]>([]);
  const [batchSize, setBatchSize] = useState(100);
  const [flushIntervalMs, setFlushIntervalMs] = useState(1000);
  const [filterExpression, setFilterExpression] = useState('');

  // Queries
  const { data: configData } = useConfig();
  const createFlowMutation = useCreateFlow();

  const { data: sourceTablesData, isLoading: sourceTablesLoading } = useDiscoverSourceTables(
    selectedSourceId,
  );
  const { data: sinkTablesData, isLoading: sinkTablesLoading } = useDiscoverSinkTables(
    selectedSinkId,
  );

  // Sync types & names
  const selectedSource = useMemo(() => {
    return configData?.config?.sources?.find((s) => s.instance_id === selectedSourceId);
  }, [configData, selectedSourceId]);

  const selectedSink = useMemo(() => {
    return configData?.config?.sinks?.find((s) => s.instance_id === selectedSinkId);
  }, [configData, selectedSinkId]);

  // Find column info for selected tables
  const sourceTableColumns = useMemo(() => {
    if (!sourceTablesData?.tables || !selectedSourceTable) return [];
    const tInfo = sourceTablesData.tables.find(
      (t) => `${t.schema}.${t.name}` === selectedSourceTable || t.name === selectedSourceTable,
    );
    return tInfo?.columns || [];
  }, [sourceTablesData, selectedSourceTable]);

  const sinkTableColumns = useMemo(() => {
    if (!sinkTablesData?.tables || !selectedSinkTable) return [];
    const tInfo = sinkTablesData.tables.find(
      (t) => `${t.schema}.${t.name}` === selectedSinkTable || t.name === selectedSinkTable,
    );
    return tInfo?.columns || [];
  }, [sinkTablesData, selectedSinkTable]);

  // Auto-generate name
  useEffect(() => {
    if (selectedSourceTable && selectedSinkTable) {
      const srcName = selectedSource?.name || selectedSourceId.substring(0, 6);
      const sinkName = selectedSink?.name || selectedSinkId.substring(0, 6);
      setFlowName(`sync-${srcName}-${sinkName}`);
    }
  }, [selectedSourceTable, selectedSinkTable, selectedSource, selectedSink, selectedSourceId, selectedSinkId]);

  // Initialize Column Mappings
  useEffect(() => {
    if (sourceTableColumns.length > 0) {
      // Auto mapping by name
      const mappings: ColumnMapping[] = sourceTableColumns.map((srcCol) => {
        // Find matching column in sink table
        const matchingSinkCol = sinkTableColumns.find(
          (sc) => sc.name.toLowerCase() === srcCol.name.toLowerCase(),
        );

        return {
          source_column: srcCol.name,
          sink_column: matchingSinkCol ? matchingSinkCol.name : srcCol.name,
          source_type: srcCol.type,
          sink_type: matchingSinkCol ? matchingSinkCol.type : srcCol.type,
          enabled: true,
        };
      });
      setColumnMappings(mappings);
    } else {
      setColumnMappings([]);
    }
  }, [sourceTableColumns, sinkTableColumns]);

  // Validate current step
  const isStepValid = useMemo(() => {
    if (step === 1) {
      return !!selectedSourceId && !!selectedSinkId;
    }
    if (step === 2) {
      return !!selectedSourceTable && !!selectedSinkTable;
    }
    if (step === 3) {
      // At least one enabled column mapping
      return columnMappings.some((m) => m.enabled);
    }
    return true;
  }, [step, selectedSourceId, selectedSinkId, selectedSourceTable, selectedSinkTable, columnMappings]);

  // Reset form helper
  const resetForm = () => {
    setStep(1);
    setFlowName('');
    setSelectedSourceId('');
    setSelectedSinkId('');
    setSelectedSourceTable('');
    setSelectedSinkTable('');
    setColumnMappings([]);
    setBatchSize(100);
    setFlushIntervalMs(1000);
    setFilterExpression('');
  };

  const handleNext = () => {
    if (isStepValid) setStep((s) => s + 1);
  };

  const handleBack = () => {
    setStep((s) => Math.max(1, s - 1));
  };

  const handleCreate = async () => {
    try {
      // Collect mappings
      const enabledMappings = columnMappings.filter((m) => m.enabled);

      await createFlowMutation.mutateAsync({
        name: flowName || 'sync-flow',
        source_id: selectedSourceId,
        sink_id: selectedSinkId,
        source_table: selectedSourceTable,
        sink_table: selectedSinkTable,
        column_mappings: enabledMappings,
        options: {
          batch_size: batchSize,
          flush_interval_ms: flushIntervalMs,
          filter_expression: filterExpression || undefined,
        },
      });

      toast.success(t('common.success'));
      resetForm();
      onOpenChange(false);
    } catch (err: any) {
      toast.error(err.message || t('manager.flows.createFailed'));
    }
  };

  // Determine if there are incompatible mappings
  const incompatibleCount = useMemo(() => {
    const sinkConnectorType = selectedSink?.type || 'stdout';
    return columnMappings.filter(
      (m) => m.enabled && !isTypeCompatible(sinkConnectorType, m.source_type, m.sink_type),
    ).length;
  }, [columnMappings, selectedSink]);

  return (
    <Dialog open={open} onOpenChange={(val) => {
      onOpenChange(val);
      if (!val) resetForm();
    }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader className="border-b pb-3">
          <DialogTitle className="flex items-center gap-2 text-foreground font-bold">
            <FolderSync className="h-5 w-5 text-sky-400" />
            {t('manager.flows.create')}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {t('manager.flows.createDesc')}
          </DialogDescription>
        </DialogHeader>

        {/* Wizard Steps indicator */}
        <div className="flex items-center justify-between px-6 py-2 bg-muted/30 rounded-lg border border-border text-xs font-semibold select-none">
          {[1, 2, 3, 4].map((s) => (
            <div key={s} className="flex items-center gap-2">
              <span className={`h-5 w-5 flex items-center justify-center rounded-full text-[10px] ${
                step === s 
                  ? 'bg-sky-500 text-slate-950 font-bold' 
                  : step > s 
                    ? 'bg-sky-500/10 text-sky-400 border border-sky-500/20' 
                    : 'bg-muted text-muted-foreground border border-border'
              }`}>
                {s}
              </span>
              <span className={step === s ? 'text-foreground' : 'text-muted-foreground'}>
                {s === 1 && t('manager.flows.steps.connectors')}
                {s === 2 && t('manager.flows.steps.tables')}
                {s === 3 && t('manager.flows.steps.columns')}
                {s === 4 && t('manager.flows.steps.options')}
              </span>
              {s < 4 && <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/50" />}
            </div>
          ))}
        </div>

        {/* Step Contents */}
        <div className="py-4 min-h-[300px] overflow-y-auto max-h-[400px] pr-1">
          {/* STEP 1: Connectors selection */}
          {step === 1 && (
            <div className="space-y-5">
              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                  {t('manager.flows.fields.selectSource')}
                </label>
                <Select value={selectedSourceId} onValueChange={(val) => setSelectedSourceId(val || '')}>
                  <SelectTrigger className="w-full h-10 text-xs">
                    <SelectValue placeholder={t('manager.flows.placeholders.chooseSource')} />
                  </SelectTrigger>
                  <SelectContent>
                    {configData?.config?.sources?.map((s) => (
                      <SelectItem key={s.instance_id} value={s.instance_id} className="text-xs">
                        {s.name || s.instance_id} ({s.type}) - {s.host}:{s.port}/{s.database}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                  {t('manager.flows.fields.selectSink')}
                </label>
                <Select value={selectedSinkId} onValueChange={(val) => setSelectedSinkId(val || '')}>
                  <SelectTrigger className="w-full h-10 text-xs">
                    <SelectValue placeholder={t('manager.flows.placeholders.chooseSink')} />
                  </SelectTrigger>
                  <SelectContent>
                    {configData?.config?.sinks?.map((s) => (
                      <SelectItem key={s.instance_id} value={s.instance_id} className="text-xs">
                        {s.name || s.instance_id} ({s.type})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {/* STEP 2: Table mappings */}
          {step === 2 && (
            <div className="space-y-5">
              <div className="grid grid-cols-2 gap-4">
                {/* Source Table */}
                <div>
                  <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                    {t('manager.flows.fields.sourceTable')}
                  </label>
                  {sourceTablesLoading ? (
                    <div className="h-10 w-full bg-muted animate-pulse rounded-lg border border-border" />
                  ) : (
                    <Select value={selectedSourceTable} onValueChange={(val) => setSelectedSourceTable(val || '')}>
                      <SelectTrigger className="w-full h-10 text-xs">
                        <SelectValue placeholder={t('manager.flows.placeholders.chooseTable')} />
                      </SelectTrigger>
                      <SelectContent>
                        {sourceTablesData?.tables?.map((tInfo) => {
                          const fullName = `${tInfo.schema}.${tInfo.name}`;
                          return (
                            <SelectItem key={fullName} value={fullName} className="text-xs">
                              {fullName}
                            </SelectItem>
                          );
                        })}
                      </SelectContent>
                    </Select>
                  )}
                </div>

                {/* Sink Table */}
                <div>
                  <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                    {t('manager.flows.fields.targetTableIndex')}
                  </label>
                  {selectedSink?.type === 'elasticsearch' ? (
                    <Input
                      value={selectedSinkTable}
                      onChange={(e) => setSelectedSinkTable(e.target.value)}
                      placeholder={t('manager.flows.placeholders.targetIndex')}
                      className="h-10 text-xs"
                    />
                  ) : selectedSink?.type === 'webhook' ? (
                    <Input
                      disabled
                      placeholder={t('manager.flows.placeholders.webhookDirect')}
                      className="h-10 text-xs cursor-not-allowed"
                    />
                  ) : sinkTablesLoading ? (
                    <div className="h-10 w-full bg-muted animate-pulse rounded-lg border border-border" />
                  ) : (
                    <Select value={selectedSinkTable} onValueChange={(val) => setSelectedSinkTable(val || '')}>
                      <SelectTrigger className="w-full h-10 text-xs">
                        <SelectValue placeholder={t('manager.flows.placeholders.chooseTable')} />
                      </SelectTrigger>
                      <SelectContent>
                        {sinkTablesData?.tables?.map((tInfo) => {
                          const fullName = `${tInfo.schema}.${tInfo.name}`;
                          return (
                            <SelectItem key={fullName} value={fullName} className="text-xs">
                              {fullName}
                            </SelectItem>
                          );
                        })}
                        {/* Allow custom fallback typing if not discovered */}
                        <SelectItem value="custom_input" className="text-xs">
                          {t('manager.flows.placeholders.customTable')}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                </div>
              </div>

              {/* Custom input for sink table fallback */}
              {selectedSinkTable === 'custom_input' && (
                <div>
                  <label className="text-[10px] uppercase tracking-wider font-semibold text-muted-foreground mb-1.5 block">
                    {t('manager.flows.fields.customTargetTable')}
                  </label>
                  <Input
                    placeholder={t('manager.flows.placeholders.customTargetTable')}
                    onChange={(e) => setSelectedSinkTable(e.target.value)}
                    className="h-10 text-xs"
                  />
                </div>
              )}
            </div>
          )}

          {/* STEP 3: Column Mappings */}
          {step === 3 && (
            <div className="space-y-4">
              {incompatibleCount > 0 && (
                <div className="rounded-xl border border-yellow-500/20 bg-yellow-500/5 p-4 flex gap-3 text-xs text-yellow-400">
                  <AlertTriangle className="h-4 w-4 shrink-0 text-yellow-400 mt-0.5" />
                  <div>
                    <span className="font-semibold text-yellow-300">{t('manager.flows.validation.incompatible')}</span>
                    <p className="opacity-95 text-[11px] leading-relaxed mt-0.5">
                      {t('manager.flows.validation.warningDesc')}
                    </p>
                  </div>
                </div>
              )}

              {/* Columns Table */}
              <div className="rounded-lg border border-border bg-card overflow-hidden text-xs">
                <Table>
                  <TableHeader className="bg-muted/50 border-b border-border text-muted-foreground select-none font-semibold">
                    <TableRow>
                      <TableHead className="px-4 py-2 w-10">{t('manager.flows.table.active')}</TableHead>
                      <TableHead className="px-4 py-2">{t('manager.flows.table.sourceColumn')}</TableHead>
                      <TableHead className="px-4 py-2 w-8 text-center"></TableHead>
                      <TableHead className="px-4 py-2">{t('manager.flows.table.sinkColumn')}</TableHead>
                      <TableHead className="px-4 py-2 w-28">{t('manager.flows.table.validation')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody className="divide-y divide-border">
                    {columnMappings.map((m, idx) => {
                      const sinkConnectorType = selectedSink?.type || 'stdout';
                      const compatible = isTypeCompatible(sinkConnectorType, m.source_type, m.sink_type);

                      const updateMapping = (key: keyof ColumnMapping, val: any) => {
                        const next = [...columnMappings];
                        next[idx] = { ...next[idx], [key]: val };
                        setColumnMappings(next);
                      };

                      return (
                        <TableRow key={m.source_column} className={`hover:bg-muted/50 ${!m.enabled ? 'opacity-40' : ''}`}>
                          <TableCell className="px-4 py-3">
                            <Switch
                              checked={m.enabled}
                              onCheckedChange={(val) => updateMapping('enabled', val)}
                              className="h-4 w-7"
                            />
                          </TableCell>
                          <TableCell className="px-4 py-3">
                            <div className="font-mono font-medium text-foreground">{m.source_column}</div>
                            <div className="text-[10px] text-muted-foreground font-mono mt-0.5">{m.source_type}</div>
                          </TableCell>
                          <TableCell className="px-4 py-3 text-center">
                            <ArrowRight className="h-3.5 w-3.5 text-muted-foreground inline" />
                          </TableCell>
                          <TableCell className="px-4 py-3">
                            <Input
                              value={m.sink_column}
                              disabled={!m.enabled}
                              onChange={(e) => updateMapping('sink_column', e.target.value)}
                              className="h-7 px-2 font-mono text-xs focus-visible:ring-0 max-w-[150px]"
                            />
                            {/* Allow sink type edits to force resolve compatibility warning if needed */}
                            <Input
                              value={m.sink_type}
                              disabled={!m.enabled}
                              onChange={(e) => updateMapping('sink_type', e.target.value)}
                              className="h-6 px-2 font-mono text-[9px] mt-1 text-muted-foreground max-w-[120px]"
                            />
                          </TableCell>
                          <TableCell className="px-4 py-3">
                            {m.enabled ? (
                              compatible ? (
                                <span className="inline-flex items-center gap-1 text-[10px] font-semibold text-emerald-400">
                                  {t('manager.flows.validation.compatible')}
                                </span>
                              ) : (
                                <span
                                  className="inline-flex items-center gap-1 text-[10px] font-semibold text-yellow-500 cursor-help"
                                  title={t('manager.flows.validation.typeMismatch', { srcType: m.source_type, sinkType: m.sink_type })}
                                >
                                  <AlertTriangle className="h-3 w-3 shrink-0 text-yellow-500" />
                                  {t('manager.flows.validation.warning')}
                                </span>
                              )
                            ) : (
                              <span className="text-[10px] text-muted-foreground font-semibold italic">
                                {t('manager.flows.validation.disabled')}
                              </span>
                            )}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}

          {/* STEP 4: Advanced Options */}
          {step === 4 && (
            <div className="space-y-5">
              {/* Flow name */}
              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                  {t('manager.flows.fields.flowName')}
                </label>
                <Input
                  value={flowName}
                  onChange={(e) => setFlowName(e.target.value)}
                  placeholder={t('manager.flows.placeholders.flowName')}
                  className="h-10 text-xs"
                />
              </div>

              {/* Sync Rate Options */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                    {t('manager.flows.fields.batchSize')}
                  </label>
                  <Input
                    type="number"
                    value={batchSize}
                    onChange={(e) => setBatchSize(Number(e.target.value))}
                    className="h-10 text-xs"
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground mb-2 block">
                    {t('manager.flows.fields.flushInterval')}
                  </label>
                  <Input
                    type="number"
                    value={flushIntervalMs}
                    onChange={(e) => setFlushIntervalMs(Number(e.target.value))}
                    className="h-10 text-xs"
                  />
                </div>
              </div>

              {/* Filter Expression */}
              <div>
                <label className="text-xs font-semibold text-muted-foreground mb-2 block flex items-center gap-1.5">
                  {t('manager.flows.fields.filterExpression')}
                  <span className="cursor-help" title={t('manager.flows.tooltips.filterExpression')}>
                    <HelpCircle className="h-3.5 w-3.5 text-muted-foreground" />
                  </span>
                </label>
                <Input
                  value={filterExpression}
                  onChange={(e) => setFilterExpression(e.target.value)}
                  placeholder={t('manager.flows.placeholders.filterExpression')}
                  className="h-10 text-xs font-mono"
                />
              </div>
            </div>
          )}
        </div>

        {/* Dialog Actions */}
        <div className="flex items-center justify-between border-t border-border pt-3">
          <Button
            variant="outline"
            onClick={step === 1 ? () => onOpenChange(false) : handleBack}
            className="h-9 text-xs cursor-pointer"
          >
            {step === 1 ? t('common.cancel') : <><ChevronLeft className="h-4 w-4 mr-1" /> {t('common.previous')}</>}
          </Button>

          {step < 4 ? (
            <Button
              onClick={handleNext}
              disabled={!isStepValid}
              className="h-9 text-xs bg-sky-500 text-slate-950 hover:bg-sky-400 font-semibold cursor-pointer"
            >
              {t('common.continue')}
              <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          ) : (
            <Button
              onClick={handleCreate}
              disabled={createFlowMutation.isPending}
              className="h-9 text-xs bg-sky-500 text-slate-950 hover:bg-sky-400 font-semibold cursor-pointer"
            >
              {createFlowMutation.isPending ? t('manager.flows.creating') : <><Play className="h-3.5 w-3.5 mr-1" /> {t('manager.flows.start')}</>}
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
