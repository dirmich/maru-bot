import { create } from 'zustand';

interface ConfirmDialogState {
    isOpen: boolean;
    title: string;
    description: string;
    onConfirm: () => void;
    show: (title: string, description: string, onConfirm: () => void) => void;
    hide: () => void;
}

// Simple legacy hook adapter (if needed for migration)
export const useConfirmDialog = () => {
    // This is a dummy hook for compatibility if any code still uses it
    // In React SPA, prefer passing state/props or using context
    return {
        show: (t: string, d: string, c: () => void) => console.log("Confirm:", t, d)
    };
};

import {
    AlertDialog as UIAlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog"

export function ConfirmDialog({
    open,
    onOpenChange,
    title,
    description,
    onConfirm
}: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    title: string;
    description: string;
    onConfirm: () => void
}) {
    return (
        <UIAlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>{title}</AlertDialogTitle>
                    <AlertDialogDescription>
                        {description}
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>취소</AlertDialogCancel>
                    <AlertDialogAction onClick={onConfirm}>확인</AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </UIAlertDialog>
    )
}
