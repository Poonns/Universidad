public class Controlador implements FocusListener{
    private Modelo modelo;
    private Vista vista;

    public Controlador(Modelo m, Vista v){
        modelo = m;
        vista = v;

        vista.getLista().addFocusListener(this);
        vista.getField().addFocusListener(this);
        vista.getBoton().addFocusListener(this);

    }
    @Override
    public void focusGained(FocusEvent e){
        modelo.setNombreFoco(e.getSource().getClass().getName());
        System.out.println("Focus ganado "+ e.getSource().getClass().getName());
    }

    @Override 
    public void focusLost(FocusEvent e){
        System.out.println("Foco perdido por " + e.getSource().getClass().getName());
    }
}