import { useTranslation } from 'react-i18next';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { Inbox } from 'lucide-react';

export interface Column<T> {
  key: string;
  header: string;
  className?: string;
  render: (item: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  isLoading?: boolean;
  emptyMessage?: string;
  onRowClick?: (item: T) => void;
  rowKey: (item: T) => string;
}

/** Generic data table with loading skeleton and empty state. */
export function DataTable<T>({
  columns,
  data,
  isLoading,
  emptyMessage,
  onRowClick,
  rowKey,
}: DataTableProps<T>) {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Inbox className="mb-3 h-10 w-10" />
        <p className="text-sm">{emptyMessage || t('common.noData')}</p>
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((col) => (
            <TableHead key={col.key} className={col.className}>
              {col.header}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.map((item) => (
          <TableRow
            key={rowKey(item)}
            onClick={() => onRowClick?.(item)}
            className={onRowClick ? 'cursor-pointer hover:bg-accent/50' : ''}
          >
            {columns.map((col) => (
              <TableCell key={col.key} className={col.className}>
                {col.render(item)}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
