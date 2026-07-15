import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import TopNav from './TopNav.svelte';

const mockGoto = vi.fn();

vi.mock('$app/navigation', () => ({
    goto: (url: string) => mockGoto(url)
}));

describe('TopNav', () => {
    it('renders and handles sign out', async () => {
        // Set mock token
        localStorage.setItem('github_token', 'mock_token');

        render(TopNav);

        const signoutBtn = screen.getByRole('button', { name: 'Sign Out' });
        expect(signoutBtn).toBeInTheDocument();

        await fireEvent.click(signoutBtn);

        // Assert token removed and goto called
        expect(localStorage.getItem('github_token')).toBeNull();
        expect(mockGoto).toHaveBeenCalledWith('/onboarding');
    });
});
