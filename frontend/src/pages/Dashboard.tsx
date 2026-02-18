// Dashboard page with stat cards and charts.

import { useQuery } from '@tanstack/react-query';
import { getStats } from '../api/dashboard';
import { getDepartments } from '../api/departments';
import { getDevices } from '../api/devices';
import { Link } from 'react-router-dom';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { PageHeader, Card, CardContent } from '../components/ui';

const CHART_COLORS = {
  accent: '#ea580c',
  success: '#22c55e',
  warning: '#f59e0b',
  muted: '#64748b',
  info: '#06b6d4',
};

export default function Dashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: getStats,
  });

  const { data: deptData } = useQuery({
    queryKey: ['departments'],
    queryFn: getDepartments,
  });

  const { data: devicesData } = useQuery({
    queryKey: ['devices', '', '', '', '', 'hostname', 'asc', 1],
    queryFn: () => getDevices({ page: 1, limit: 1000 }),
  });

  if (isLoading) {
    return (
      <div className="space-y-6 animate-fade-in">
        <div className="h-8 w-48 bg-bg-tertiary rounded-lg animate-pulse" />
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-28 bg-bg-secondary rounded-xl border border-border-primary animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return <p className="text-danger">Failed to load dashboard data.</p>;
  }

  const total = data?.total ?? 0;
  const online = data?.online ?? 0;
  const offline = data?.offline ?? 0;
  const inactive = data?.inactive ?? 0;

  // Data for the status pie chart
  const pieData = [
    { name: 'Online', value: online, color: CHART_COLORS.success },
    { name: 'Offline', value: offline, color: CHART_COLORS.muted },
    { name: 'Inactive', value: inactive, color: CHART_COLORS.warning },
  ].filter((d) => d.value > 0);

  // Data for the department bar chart
  const devices = devicesData?.devices ?? [];
  const departments = deptData?.departments ?? [];
  const deptCounts = departments.map((d) => ({
    name: d.name.length > 12 ? d.name.substring(0, 12) + '…' : d.name,
    count: devices.filter((dev) => dev.department_id === d.id).length,
  }));
  const unassigned = devices.filter((dev) => !dev.department_id).length;
  if (unassigned > 0) {
    deptCounts.push({ name: 'Unassigned', count: unassigned });
  }

  const stats = [
    {
      label: 'Total Devices',
      value: total,
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z" />
        </svg>
      ),
      gradient: 'from-accent/15 to-accent/5',
      iconColor: 'text-accent',
      borderColor: 'border-accent/20',
    },
    {
      label: 'Online',
      value: online,
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
      gradient: 'from-success/15 to-success/5',
      iconColor: 'text-success',
      borderColor: 'border-success/20',
    },
    {
      label: 'Offline',
      value: offline,
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
        </svg>
      ),
      gradient: 'from-bg-tertiary to-bg-secondary',
      iconColor: 'text-text-muted',
      borderColor: 'border-border-primary',
    },
    {
      label: 'Inactive',
      value: inactive,
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
        </svg>
      ),
      gradient: 'from-warning/15 to-warning/5',
      iconColor: 'text-warning',
      borderColor: 'border-warning/20',
    },
  ];

  return (
    <div className="animate-fade-in">
      <PageHeader title="Dashboard" subtitle={`${total} devices registered`} />

      {/* Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {stats.map((s, i) => (
          <div
            key={s.label}
            className={`bg-gradient-to-br ${s.gradient} rounded-xl border ${s.borderColor} p-5 animate-slide-up`}
            style={{ animationDelay: `${i * 50}ms`, animationFillMode: 'both' }}
          >
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm font-medium text-text-secondary">{s.label}</span>
              <div className={`${s.iconColor} opacity-70`}>{s.icon}</div>
            </div>
            <p className="text-3xl font-bold text-text-primary">{s.value}</p>
          </div>
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Status Distribution Pie */}
        {pieData.length > 0 && (
          <Card>
            <CardContent>
              <h3 className="text-sm font-semibold text-text-primary mb-4">Status Distribution</h3>
              <div className="flex items-center justify-center">
                <ResponsiveContainer width="100%" height={220}>
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      innerRadius={55}
                      outerRadius={85}
                      paddingAngle={4}
                      dataKey="value"
                      stroke="none"
                    >
                      {pieData.map((entry, index) => (
                        <Cell key={index} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--color-bg-secondary)',
                        borderColor: 'var(--color-border-primary)',
                        borderRadius: '0.75rem',
                        fontSize: '0.75rem',
                        color: 'var(--color-text-primary)',
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="flex justify-center gap-6 mt-2">
                {pieData.map((d) => (
                  <div key={d.name} className="flex items-center gap-2 text-xs text-text-secondary">
                    <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: d.color }} />
                    {d.name} ({d.value})
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Devices by Department */}
        {deptCounts.length > 0 && (
          <Card>
            <CardContent>
              <h3 className="text-sm font-semibold text-text-primary mb-4">Devices by Department</h3>
              <ResponsiveContainer width="100%" height={220}>
                <BarChart data={deptCounts} margin={{ top: 5, right: 10, bottom: 5, left: -10 }}>
                  <XAxis
                    dataKey="name"
                    tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }}
                    axisLine={false}
                    tickLine={false}
                    allowDecimals={false}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--color-bg-secondary)',
                      borderColor: 'var(--color-border-primary)',
                      borderRadius: '0.75rem',
                      fontSize: '0.75rem',
                      color: 'var(--color-text-primary)',
                    }}
                    cursor={{ fill: 'var(--color-bg-tertiary)', opacity: 0.5 }}
                  />
                  <Bar
                    dataKey="count"
                    fill={CHART_COLORS.accent}
                    radius={[6, 6, 0, 0]}
                    maxBarSize={40}
                  />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        )}
      </div>

      {total > 0 && (
        <Link
          to="/devices"
          className="inline-flex items-center gap-1.5 text-sm text-accent hover:text-accent-hover font-medium transition-colors"
        >
          View all devices
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
          </svg>
        </Link>
      )}
    </div>
  );
}
