// Client-side dialog placeholder (Mocking missing alert-dialog)
export const useConfirmDialog = () => {
    return {
        show: (t: string, d: string, c: () => void) => {
            if (window.confirm(`${t}\n\n${d}`)) {
                c();
            }
        }
    };
};

export function ConfirmDialog({
    open,
    onOpenChange,
    title,
    description,
    onConfirm
}: any) {
    if (!open) return null;
    return (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
            <div style={{ background: 'white', padding: '20px', borderRadius: '8px', color: 'black' }}>
                <h2>{title}</h2>
                <p>{description}</p>
                <div style={{ display: 'flex', gap: '10px', marginTop: '20px' }}>
                    <button onClick={() => onOpenChange(false)}>Cancel</button>
                    <button onClick={() => { onConfirm(); onOpenChange(false); }}>Confirm</button>
                </div>
            </div>
        </div>
    )
}
