import { useState } from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Cpu, Save, Plus, Trash } from 'lucide-react';
import { toast } from 'sonner';

// Placeholder for GpioSchematic since we're migrating
const GpioSchematic = ({ configuredPins }: { configuredPins: number[] }) => (
    <div className="bg-slate-100 p-8 rounded-lg text-center border-2 border-dashed border-slate-300">
        <h3 className="font-semibold text-slate-500 mb-2">Raspberry Pi GPIO Header</h3>
        <p className="text-xs text-slate-400">Visualization Component Coming Soon</p>
        <div className="mt-4 grid grid-cols-2 gap-2 max-w-xs mx-auto text-xs font-mono">
            {[...Array(40)].map((_, i) => (
                <div key={i} className={`h-4 w-4 rounded-full mx-auto ${configuredPins.includes(i + 1) ? 'bg-orange-500' : 'bg-slate-300'}`}></div>
            ))}
        </div>
    </div>
);

interface PinConfig {
    pin: number;
    mode: string;
    label: string;
}

export function GpioPage() {
    const [configuredPins, setConfiguredPins] = useState<PinConfig[]>([
        { pin: 7, mode: 'OUT', label: 'Status LED' },
        { pin: 11, mode: 'IN', label: 'Button 1' },
    ]);

    const handleAddPin = () => {
        // Find next available pin or just use 0 as placeholder
        const newPin = { pin: 0, mode: 'OUT', label: 'New Device' };
        setConfiguredPins([...configuredPins, newPin]);
    };

    const handleRemovePin = (index: number) => {
        const newPins = [...configuredPins];
        newPins.splice(index, 1);
        setConfiguredPins(newPins);
    };

    const handleUpdatePin = (index: number, field: keyof PinConfig, value: any) => {
        const newPins = [...configuredPins];
        (newPins[index] as any)[field] = value;
        setConfiguredPins(newPins);
    }

    const handleSave = () => {
        console.log("Saving GPIO config:", configuredPins);
        toast.success('GPIO 설정이 저장되었습니다.');
    };

    return (
        <div className="p-6 max-w-6xl mx-auto space-y-6">
            <header className="mb-6">
                <h1 className="text-2xl font-bold flex items-center gap-2">
                    <Cpu className="text-orange-600" /> GPIO 제어 및 설정
                </h1>
                <p className="text-sm text-slate-500">Raspberry Pi의 핀 맵을 시각적으로 확인하고 하드웨어 인터페이스를 설정합니다.</p>
            </header>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="space-y-6">
                    <Card className="border-none shadow-lg">
                        <CardHeader>
                            <CardTitle className="text-lg">핀 맵 스케매틱</CardTitle>
                            <CardDescription>핀 번호를 클릭하여 상세 정보를 확인하세요.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <GpioSchematic configuredPins={configuredPins.map(p => p.pin)} />
                        </CardContent>
                    </Card>
                </div>

                <div className="space-y-6 flex flex-col">
                    <Card className="border-none shadow-lg flex-1">
                        <CardHeader className="flex flex-row items-center justify-between">
                            <div>
                                <CardTitle className="text-lg">설정된 장치</CardTitle>
                                <CardDescription>활성화된 GPIO 핀 목록입니다.</CardDescription>
                            </div>
                            <Button size="sm" onClick={handleAddPin} className="bg-orange-600 hover:bg-orange-700 text-white">
                                <Plus className="w-4 h-4 mr-1" /> 추가
                            </Button>
                        </CardHeader>
                        <CardContent>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Pin</TableHead>
                                        <TableHead>Mode</TableHead>
                                        <TableHead>Label</TableHead>
                                        <TableHead className="w-10"></TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {configuredPins.map((item, idx) => (
                                        <TableRow key={idx}>
                                            <TableCell className="font-mono">
                                                <input
                                                    type="number"
                                                    className="w-12 bg-transparent border-b border-dashed text-center"
                                                    value={item.pin}
                                                    onChange={(e) => handleUpdatePin(idx, 'pin', parseInt(e.target.value))}
                                                />
                                            </TableCell>
                                            <TableCell>
                                                <Select
                                                    value={item.mode}
                                                    onValueChange={(v) => handleUpdatePin(idx, 'mode', v)}
                                                >
                                                    <SelectTrigger className="w-24 h-8 text-xs">
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value="OUT">OUT</SelectItem>
                                                        <SelectItem value="IN">IN</SelectItem>
                                                        <SelectItem value="PWM">PWM</SelectItem>
                                                        <SelectItem value="I2C">I2C</SelectItem>
                                                        <SelectItem value="SPI">SPI</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </TableCell>
                                            <TableCell>
                                                <input
                                                    className="bg-transparent border-b border-dashed border-slate-300 dark:border-slate-700 focus:outline-none focus:border-orange-500 w-full text-sm"
                                                    value={item.label}
                                                    onChange={(e) => handleUpdatePin(idx, 'label', e.target.value)}
                                                />
                                            </TableCell>
                                            <TableCell>
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    className="h-8 w-8 text-slate-400 hover:text-red-500"
                                                    onClick={() => handleRemovePin(idx)}
                                                >
                                                    <Trash className="w-4 h-4" />
                                                </Button>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                    {configuredPins.length === 0 && (
                                        <TableRow>
                                            <TableCell colSpan={4} className="text-center text-slate-400 py-8">
                                                설정된 GPIO 핀이 없습니다.
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        </CardContent>
                        <CardFooter className="justify-end border-t pt-4">
                            <Button onClick={handleSave} className="bg-blue-600 hover:bg-blue-700 text-white">
                                <Save className="w-4 h-4 mr-2" /> 설정 저장
                            </Button>
                        </CardFooter>
                    </Card>
                </div>
            </div>
        </div>
    );
}
