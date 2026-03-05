import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import LoginModal from './LoginModal';

vi.mock('../api/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
}));

import { login, register } from '../api/auth';
const mockLogin = vi.mocked(login);
const mockRegister = vi.mocked(register);

const mockOnAuth = vi.fn();
const mockOnClose = vi.fn();

function renderModal() {
  return render(<LoginModal onAuth={mockOnAuth} onClose={mockOnClose} />);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('LoginModal', () => {
  it('renderiza el formulario de login por defecto', () => {
    renderModal();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('Iniciar sesión')).toBeInTheDocument();
    expect(screen.getByLabelText('Usuario')).toBeInTheDocument();
    expect(screen.getByLabelText('Contraseña')).toBeInTheDocument();
  });

  it('cambia al modo registro al pulsar "Registrarse"', () => {
    renderModal();
    fireEvent.click(screen.getByRole('button', { name: 'Registrarse' }));
    expect(screen.getByRole('heading', { name: 'Crear cuenta' })).toBeInTheDocument();
  });

  it('llama a onClose al pulsar el botón cerrar', () => {
    renderModal();
    fireEvent.click(screen.getByRole('button', { name: 'Cerrar' }));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('llama a onClose al hacer click en el overlay', () => {
    renderModal();
    fireEvent.click(screen.getByRole('dialog'));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('llama a login con las credenciales del formulario', async () => {
    mockLogin.mockResolvedValue({ token: 'tok', user: { userId: 1, username: 'carlos' } });
    const { container } = renderModal();
    fireEvent.change(screen.getByLabelText('Usuario'), { target: { value: 'carlos' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'password123' } });
    fireEvent.click(container.querySelector('button[type="submit"]')!);
    await waitFor(() => expect(mockLogin).toHaveBeenCalledWith('carlos', 'password123'));
  });

  it('llama a onAuth con el token y el usuario tras login exitoso', async () => {
    mockLogin.mockResolvedValue({ token: 'tok123', user: { userId: 2, username: 'carlos' } });
    const { container } = renderModal();
    fireEvent.change(screen.getByLabelText('Usuario'), { target: { value: 'carlos' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'password123' } });
    fireEvent.click(container.querySelector('button[type="submit"]')!);
    await waitFor(() =>
      expect(mockOnAuth).toHaveBeenCalledWith({
        token: 'tok123',
        user: { userId: 2, username: 'carlos' },
      }),
    );
  });

  it('muestra el error cuando el login falla', async () => {
    mockLogin.mockRejectedValue(new Error('Usuario o contraseña incorrectos'));
    const { container } = renderModal();
    fireEvent.change(screen.getByLabelText('Usuario'), { target: { value: 'carlos' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'wrongpass' } });
    fireEvent.click(container.querySelector('button[type="submit"]')!);
    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Usuario o contraseña incorrectos'),
    );
  });

  it('llama a register en modo registro', async () => {
    mockRegister.mockResolvedValue({ token: 'tok', user: { userId: 3, username: 'nuevo' } });
    const { container } = renderModal();
    fireEvent.click(screen.getByRole('button', { name: 'Registrarse' }));
    fireEvent.change(screen.getByLabelText('Usuario'), { target: { value: 'nuevo' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'password123' } });
    fireEvent.click(container.querySelector('button[type="submit"]')!);
    await waitFor(() => expect(mockRegister).toHaveBeenCalledWith('nuevo', 'password123'));
  });

  it('deshabilita el botón mientras carga', async () => {
    mockLogin.mockImplementation(() => new Promise(() => {})); // never resolves
    const { container } = renderModal();
    fireEvent.change(screen.getByLabelText('Usuario'), { target: { value: 'carlos' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'password123' } });
    fireEvent.click(container.querySelector('button[type="submit"]')!);
    await waitFor(() =>
      expect(container.querySelector('button[type="submit"]')).toBeDisabled(),
    );
  });

  it('limpia el error al cambiar de pestaña', async () => {
    mockLogin.mockRejectedValue(new Error('Error de prueba'));
    const { container } = renderModal();
    fireEvent.change(screen.getByLabelText('Usuario'), { target: { value: 'u' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'p' } });
    fireEvent.click(container.querySelector('button[type="submit"]')!);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'Registrarse' }));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
