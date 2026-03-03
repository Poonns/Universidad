public class Controlador implements MouseListener{
    private Vista vista;

    public Controlador(Vista v){
        vista = v;

        vista.getPanel().addMouseListener(new MouseAdapter() {
            @Override
            public void mousePressed(MouseEvent e){
                int x = e.getX();
                int y = e.getY();

                vista.setPosRaton(x,y);
            }
        });
    }
}