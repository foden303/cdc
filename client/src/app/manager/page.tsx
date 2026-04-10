"use client";

import { useEffect, useState } from "react";
import { 
  getConfigAction, 
  addSourceAction, 
  removeSourceAction, 
  updateSourceAction,
  addSinkAction,
  removeSinkAction,
  updateSinkAction
} from "@/lib/actions";
import { AppConfig, SourceConfig, SinkConfig } from "@/lib/grpc";
import { 
  Plus, 
  Settings, 
  Trash2, 
  Database, 
  Box, 
  Activity, 
  CheckCircle2, 
  AlertCircle,
  X,
  Save,
  ChevronRight,
  Server,
  Layers,
  Zap,
  MoreVertical
} from "lucide-react";

export default function ManagerPage() {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalType, setModalType] = useState<"source" | "sink">("source");
  const [editingItem, setEditingItem] = useState<any>(null);
  const [selectedType, setSelectedType] = useState<string>("");
  const [isSaving, setIsSaving] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);

  useEffect(() => {
    fetchConfig();
  }, []);
  
  const fetchConfig = async () => {
    setLoading(true);
    const res = await getConfigAction() as { config: AppConfig };
    if (res?.config) {
      setConfig(res.config);
    }
    setLoading(false);
  };

  const handleRemove = async (type: "source" | "sink", id: string) => {
    if (!confirm(`Are you sure you want to remove this ${type}?`)) return;
    if (type === "source") await removeSourceAction(id);
    else await removeSinkAction(id);
    fetchConfig();
  };
  
  const openAddModal = (type: "source" | "sink") => {
    setModalType(type);
    setEditingItem(null);
    setSelectedType(type === "source" ? "postgres" : "clickhouse");
    setShowAdvanced(false);
    setIsModalOpen(true);
  };

  const openEditModal = (type: "source" | "sink", item: any) => {
    setModalType(type);
    setEditingItem(item);
    setSelectedType(item.type);
    setShowAdvanced(false);
    setIsModalOpen(true);
  };

  const handleSave = async (e: any) => {
    e.preventDefault();
    setIsSaving(true);
    const formData = new FormData(e.target);
    const data: any = {};
    
    formData.forEach((value, key) => {
      if (key === 'tables' || key === 'value_fields') {
        data[key] = (value as string).split(',').map(t => t.trim()).filter(Boolean);
      } else if (key === 'url') {
        data[key] = [(value as string)].filter(Boolean);
      } else if (key.startsWith('redis_')) {
        if (!data.redis) data.redis = {};
        const field = key.replace('redis_', '');
        if (field === 'ttl') data.redis[field] = parseInt(value as string) || 0;
        else data.redis[field] = value;
      } else if (key === 'field_mapping') {
        try {
          data[key] = JSON.parse(value as string);
        } catch {
          // Fallback if not JSON
        }
      } else if (key === 'port' || key === 'batch_size' || key === 'flush_interval_ms' || key === 'max_retries') {
        const num = parseInt(value as string);
        if (!isNaN(num)) data[key] = num;
      } else if (value) {
        data[key] = value;
      }
    });

    try {
      if (modalType === "source") {
        if (editingItem) await updateSourceAction({ ...editingItem, ...data });
        else await addSourceAction(data);
      } else {
        if (editingItem) await updateSinkAction({ ...editingItem, ...data });
        else await addSinkAction(data);
      }
      setIsModalOpen(false);
      fetchConfig();
    } catch (err) {
      console.error(err);
      alert("Failed to save configuration.");
    } finally {
      setIsSaving(false);
    }
  };

  if (loading && !config) return <LoadingState />;

  return (
    <>
      <MainView 
        config={config} 
        openAddModal={openAddModal} 
        openEditModal={openEditModal} 
        handleRemove={handleRemove} 
      />

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-6 backdrop-blur-sm bg-black/40 animate-in fade-in duration-300">
           <div className="w-full max-w-2xl bg-slate-900 border border-white/10 rounded-[2.5rem] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-300">
              <div className="flex items-center justify-between p-8 border-b border-white/5 bg-white/[0.02]">
                 <div className="flex items-center gap-4">
                    <div className={`p-3 rounded-2xl ${modalType === 'source' ? 'bg-blue-500/10 text-blue-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                       {modalType === 'source' ? <Database /> : <Box />}
                    </div>
                    <div>
                       <h3 className="text-xl font-bold text-white">{editingItem ? 'Edit Instance' : `Add New ${modalType}`}</h3>
                       <p className="text-xs text-slate-500 uppercase font-black tracking-widest mt-1">Configure {modalType} connectivity</p>
                    </div>
                 </div>
                 <button onClick={() => setIsModalOpen(false)} className="w-10 h-10 rounded-full hover:bg-white/5 flex items-center justify-center text-slate-500">
                    <X />
                 </button>
              </div>

              <form onSubmit={handleSave} className="p-8 space-y-6 max-h-[75vh] overflow-y-auto custom-scrollbar">
                 <div className="grid grid-cols-2 gap-6">
                    <InputGroup label="Display Name" name="name" defaultValue={editingItem?.name} placeholder="e.g. Production Orders" />
                    <InputGroup label="Target Topic" name="topic" defaultValue={editingItem?.topic} placeholder="e.g. cdc.orders" />
                 </div>
                 
                 <div className="p-4 rounded-2xl bg-blue-500/5 border border-blue-500/10 mb-6 font-mono text-[10px]">
                    <div className="flex gap-3">
                       <Activity className="w-4 h-4 text-blue-400 mt-1 shrink-0" />
                       <div className="text-slate-400 leading-relaxed uppercase tracking-widest font-black">
                          <strong className="text-blue-300 block mb-1">Routing Strategy</strong>
                          {modalType === 'source' ? (
                            <>PRODUCING TO: <span className="text-blue-400">topic.[id].[schema].[table]</span></>
                          ) : (
                            <>CONSUMING FROM: <span className="text-emerald-400">pattern.*</span></>
                          )}
                       </div>
                    </div>
                 </div>

                 <div className="grid grid-cols-2 gap-6">
                    <div className="space-y-2">
                       <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Instance Type</label>
                        <select 
                          name="type" 
                          value={selectedType}
                          onChange={(e) => setSelectedType(e.target.value)}
                          className="w-full bg-white/[0.03] border border-white/10 rounded-2xl px-4 py-3 text-white outline-none focus:border-blue-500/50 transition-all appearance-none"
                        >
                           {modalType === 'source' ? (
                             <>
                               <option value="postgres">PostgreSQL (Logical Replication)</option>
                               <option value="mysql">MySQL (BinLog)</option>
                             </>
                           ) : (
                             <>
                               <option value="clickhouse">Clickhouse DB</option>
                               <option value="elasticsearch">Elasticsearch</option>
                               <option value="redis">Redis Cache</option>
                             </>
                           )}
                        </select>
                    </div>
                    <InputGroup label="Host / Endpoint" name={modalType === 'source' ? 'host' : 'url'} defaultValue={modalType === 'source' ? editingItem?.host : editingItem?.url?.[0]} placeholder="e.g. 10.0.0.1" />
                 </div>

                 {modalType === 'source' ? (
                   <>
                    <div className="grid grid-cols-2 gap-6">
                       <InputGroup label="Port" name="port" type="number" defaultValue={editingItem?.port || 5432} />
                       <InputGroup label="Database" name="database" defaultValue={editingItem?.database} placeholder="e.g. orders_db" />
                    </div>
                    <div className="grid grid-cols-2 gap-6">
                       <InputGroup label="Username" name="username" defaultValue={editingItem?.username} />
                       <InputGroup label="Password" name="password" type="password" defaultValue={editingItem?.password} />
                    </div>
                    <InputGroup label="Capture Tables" name="tables" defaultValue={editingItem?.tables?.join(', ')} placeholder="public.orders, public.details" />
                   </>
                 ) : (
                   <>
                     <div className="grid grid-cols-2 gap-6">
                        <InputGroup label="Batch Size" name="batch_size" type="number" defaultValue={editingItem?.batch_size || 500} />
                        <InputGroup label="Index / Table" name="index" defaultValue={editingItem?.index} placeholder="e.g. records_index" />
                     </div>
                     <div className="grid grid-cols-2 gap-6">
                        <InputGroup label="Username" name="username" defaultValue={editingItem?.username} />
                        <InputGroup label="Password" name="password" type="password" defaultValue={editingItem?.password} />
                     </div>
                     <div className="space-y-2">
                        <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Field Mapping (JSON)</label>
                        <textarea 
                          name="field_mapping" 
                          defaultValue={editingItem?.field_mapping ? JSON.stringify(editingItem.field_mapping) : ''} 
                          placeholder='{"id": "user_id", "email": "contact_email"}'
                          className="w-full bg-white/[0.03] border border-white/10 rounded-2xl px-4 py-3 text-white outline-none focus:border-blue-500/50 transition-all font-medium h-24"
                        />
                     </div>

                     {/* Redis Specific */}
                     {selectedType === 'redis' && (
                       <div className="p-6 rounded-3xl bg-blue-500/5 border border-blue-500/10 space-y-6 mt-4">
                          <h3 className="text-xs font-black text-blue-400 uppercase tracking-widest">Redis Strategy</h3>
                          <div className="grid grid-cols-2 gap-6">
                            <div className="space-y-2">
                              <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest">Command</label>
                              <select name="redis_command" defaultValue={editingItem?.redis?.command || 'set'} className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-2.5 text-white">
                                <option value="set">SET</option>
                                <option value="hset">HSET</option>
                                <option value="sadd">SADD</option>
                                <option value="bf.add">Bloom Filter ADD</option>
                              </select>
                            </div>
                            <InputGroup label="TTL (Seconds)" name="redis_ttl" type="number" defaultValue={editingItem?.redis?.ttl || 0} />
                          </div>
                          <InputGroup label="Key Template (CEL)" name="redis_key_template" defaultValue={editingItem?.redis?.key_template} placeholder="user:profile:${id}" />
                          <InputGroup label="Value Fields (Comma Separated)" name="value_fields" defaultValue={editingItem?.redis?.value_fields?.join(', ')} placeholder="id, email, status" />
                       </div>
                     )}
                   </>
                 )}

                 <div className="pt-4">
                    <button 
                      type="button" 
                      onClick={() => setShowAdvanced(!showAdvanced)}
                      className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-white transition-colors"
                    >
                       <Settings className={`w-3 h-3 transition-transform ${showAdvanced ? 'rotate-90' : ''}`} />
                       {showAdvanced ? 'Hide Advanced Settings' : 'Show Advanced Settings'}
                    </button>
                    
                    {showAdvanced && (
                      <div className="mt-6 p-6 rounded-3xl bg-black/20 border border-white/5 space-y-6 animate-in slide-in-from-top-2 duration-300">
                         <InputGroup 
                           label="Instance ID (Unique Slug)" 
                           name="instance_id" 
                           defaultValue={editingItem?.instance_id} 
                           placeholder="Auto-generated if blank" 
                           readOnly={!!editingItem} 
                         />
                         {modalType === 'source' && (
                           <div className="grid grid-cols-2 gap-6">
                             <InputGroup label="Replication Slot" name="slot_name" defaultValue={editingItem?.slot_name} placeholder="Auto-managed" />
                             <InputGroup label="Publication" name="publication_name" defaultValue={editingItem?.publication_name} placeholder="Auto-managed" />
                           </div>
                         )}
                      </div>
                    )}
                 </div>

                 <div className="pt-6 border-t border-white/5 flex gap-4">
                    <button 
                      type="submit" 
                      disabled={isSaving}
                      className="flex-1 bg-blue-600 hover:bg-blue-500 disabled:bg-blue-900 py-4 rounded-2xl text-white font-bold flex items-center justify-center gap-2 transition-all"
                    >
                       {isSaving ? <Activity className="w-5 h-5 animate-spin" /> : <Save className="w-5 h-5" />}
                       {editingItem ? 'Update Configuration' : 'Deploy Instance'}
                    </button>
                    <button 
                      type="button" 
                      onClick={() => setIsModalOpen(false)}
                      className="px-8 py-4 rounded-2xl bg-white/5 text-slate-400 font-bold hover:bg-white/10 transition-all"
                    >
                       Cancel
                    </button>
                 </div>
              </form>
           </div>
        </div>
      )}
    </>
  );
}

function SectionHeader({ icon: Icon, title, count, color }: any) {
  const colors: any = {
    blue: "text-blue-400 bg-blue-500/10",
    emerald: "text-emerald-400 bg-emerald-500/10"
  };
  return (
    <div className="flex items-center justify-between px-2">
      <h2 className="text-2xl font-black text-white tracking-tight flex items-center gap-3">
        <Icon className={`w-6 h-6 ${colors[color].split(' ')[0]}`} />
        {title}
      </h2>
      <span className="text-[10px] font-black text-slate-500 uppercase tracking-widest">{count} TOTAL</span>
    </div>
  );
}

const LoadingState = () => (
  <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
    <div className="w-16 h-16 rounded-full border-4 border-blue-500/20 border-t-blue-500 animate-spin" />
    <p className="text-slate-500 font-bold uppercase tracking-widest text-[10px]">Syncing Orchestration State...</p>
  </div>
);

const Header = ({ onAdd }: any) => (
  <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-8">
    <div className="space-y-3">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-[10px] font-black uppercase tracking-widest">
        <Settings className="w-3 h-3" /> Orchestration Panel
      </div>
      <h1 className="text-5xl font-black text-white tracking-tighter">Pipeline Manager</h1>
      <p className="text-slate-400 text-lg max-w-xl">Deploy and manage CDC instances with real-time hot-reloading.</p>
    </div>
    <div className="flex gap-4">
      <button onClick={() => onAdd("source")} className="px-6 py-3 rounded-2xl bg-blue-600 hover:bg-blue-500 text-white font-bold text-sm flex items-center gap-2 transition-all">
        <Plus className="w-4 h-4" /> Add Source
      </button>
      <button onClick={() => onAdd("sink")} className="px-6 py-3 rounded-2xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-sm flex items-center gap-2 transition-all">
        <Plus className="w-4 h-4" /> Add Sink
      </button>
    </div>
  </div>
);

// Add the full page structure back
function MainView({ config, openAddModal, openEditModal, handleRemove }: any) {
  return (
    <div className="space-y-12 pb-20 animate-in fade-in slide-in-from-bottom-4 duration-1000">
      <Header onAdd={openAddModal} />
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-12">
        <div className="space-y-6">
          <SectionHeader icon={Database} title="Ingestion Sources" count={config?.sources?.length} color="blue" />
          <div className="space-y-4">
            {config?.sources?.map((s: any) => (
              <ConfigCard key={s.instance_id} title={s.name || s.instance_id} type={s.type} details={`${s.database} @ ${s.host}`} onEdit={() => openEditModal("source", s)} onRemove={() => handleRemove("source", s.instance_id)} color="blue" />
            ))}
          </div>
        </div>
        <div className="space-y-6">
          <SectionHeader icon={Box} title="Delivery Sinks" count={config?.sinks?.length} color="emerald" />
          <div className="space-y-4">
            {config?.sinks?.map((s: any) => (
              <ConfigCard key={s.instance_id} title={s.name || s.instance_id} type={s.type} details={s.url?.join(', ')} onEdit={() => openEditModal("sink", s)} onRemove={() => handleRemove("sink", s.instance_id)} color="emerald" />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function ConfigCard({ title, type, details, onEdit, onRemove, color }: any) {
  const colors: any = {
    blue: "from-blue-500/10 to-transparent border-blue-500/10 hover:border-blue-500/30",
    emerald: "from-emerald-500/10 to-transparent border-emerald-500/10 hover:border-emerald-500/30"
  };

  return (
    <div className={`p-6 rounded-[2rem] bg-white/[0.02] border backdrop-blur-md bg-gradient-to-br transition-all group ${colors[color]}`}>
       <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
             <div className="p-2.5 rounded-xl bg-white/[0.03] border border-white/5">
                {color === 'blue' ? <Database className="w-5 h-5 text-blue-400" /> : <Box className="w-5 h-5 text-emerald-400" />}
             </div>
             <div>
                <h4 className="font-bold text-white group-hover:text-blue-200 transition-colors">{title}</h4>
                <div className="text-[10px] uppercase font-black tracking-widest text-slate-500 mt-0.5">{type} Engine</div>
             </div>
          </div>
          <div className="flex gap-2">
             <button onClick={onEdit} className="p-2 rounded-xl bg-white/5 hover:bg-white/10 text-slate-400 hover:text-white transition-all">
                <Settings className="w-4 h-4" />
             </button>
             <button onClick={onRemove} className="p-2 rounded-xl bg-white/5 hover:bg-rose-500/20 text-slate-400 hover:text-rose-400 transition-all">
                <Trash2 className="w-4 h-4" />
             </button>
          </div>
       </div>
       <div className="text-sm text-slate-500 font-medium px-1 flex items-center gap-2">
          <Server className="w-3 h-3 text-slate-600" />
          <span className="truncate">{details}</span>
       </div>
    </div>
  );
}

function InputGroup({ label, name, defaultValue, placeholder, type = "text", readOnly = false }: any) {
  return (
    <div className="space-y-2">
       <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">{label}</label>
       <input 
         type={type} 
         name={name} 
         defaultValue={defaultValue} 
         placeholder={placeholder}
         readOnly={readOnly}
         className={`w-full bg-white/[0.03] border border-white/10 rounded-2xl px-4 py-3 text-white outline-none focus:border-blue-500/50 transition-all font-medium ${readOnly ? 'opacity-50 cursor-not-allowed' : ''}`}
       />
    </div>
  );
}
