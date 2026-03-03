public class Controlador{
    Vista vista;
    Modelo modelo;

    public Controlador(Vista v, Modelo m){
        vista = v;
        modelo = m;

        vista.getPanel().addMouseListener(new MouseAdapter(){
            @Override
            public void mousePressed(MouseEvent e){
                if(e.getButton() == MouseEvent.BUTTON1){
                    modelo.siguienteColor();
                    vista.getPanel().setColor(modelo.getColorActual());
                } else if(e.Button() == MouseEvent.BUTTON3){
                    modelo.anteriorColor();
                    vista.getPanel().setColor(modelo.getColorActual());
                }
            }
        });
    }
}